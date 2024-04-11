package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type UpdateModelRepository interface {
	ChangeFeaturesCount(ctx context.Context, userID, modelType string, faceFeaturesCount int) error
	SetModelURL(ctx context.Context, url, modelType, userID string) error
	SetModelStatus(ctx context.Context, status string, modelType string, userID string) error
}

type Consumer interface {
	Consume(queue string) (<-chan amqp.Delivery, func(), error)
}

type ModelUpdater struct {
	updateModelRepository UpdateModelRepository
	transactor            postgresql.Transactor

	consumer    Consumer
	resultQueue string

	logger *zap.Logger
}

func NewModelUpdater(
	updateModelRepository UpdateModelRepository,
	transactor postgresql.Transactor,
	consumer Consumer,
	resultQueue string,
	logger *zap.Logger,
) *ModelUpdater {
	return &ModelUpdater{
		updateModelRepository: updateModelRepository,
		transactor:            transactor,
		consumer:              consumer,
		resultQueue:           resultQueue,
		logger:                logger,
	}
}

type Message struct {
	ModelType     string `json:"model_type"`
	UserID        string `json:"user_id"`
	ModelURL      string `json:"model_url"`
	FeaturesCount int    `json:"features_count"`
}

func (m ModelUpdater) StartModelUpdate() {
	op := "model_trainer.ModelUpdater.updateModels"

	msgs, release, err := m.consumer.Consume(m.resultQueue)
	if err != nil {
		m.logger.Error(fmt.Sprintf("%s: %s", op, err.Error()))
		return
	}
	defer release()

	for msg := range msgs {
		var model Message
		err = json.Unmarshal(msg.Body, &model)
		if err != nil {
			m.logger.Error(fmt.Sprintf("%s: %s", op, err.Error()))
			continue
		}
		txErr := m.transactor.WithinTransaction(context.Background(), func(txCtx context.Context) error {
			err := m.updateModelRepository.SetModelURL(txCtx, model.ModelURL, model.ModelType, model.UserID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			err = m.updateModelRepository.SetModelStatus(txCtx, data.StatusTrained, model.ModelType, model.UserID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			err = m.updateModelRepository.ChangeFeaturesCount(txCtx, model.UserID, model.ModelType, -model.FeaturesCount)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			return nil
		})
		if txErr != nil {
			m.logger.Error(txErr.Error())
		}

		m.logger.Info(fmt.Sprintf("%s: update model finished", op))
	}
}
