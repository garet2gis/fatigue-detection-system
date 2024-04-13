package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
)

type ViewModelRepository interface {
	ViewNotLearnedModels(ctx context.Context, modelType string, trainThreshold uint64) ([]data.MLModel, error)
	ViewNotFineTunedFaceModels(ctx context.Context, modelType string, tuneThreshold uint64) ([]data.MLModel, error)
	SetModelStatus(ctx context.Context, status string, modelType string, userID string) error
}

type Producer interface {
	Publish(queue string, message []byte) error
}

type ModelTrainThreshold struct {
	TrainThreshold uint64 `json:"train_threshold"`
	TuneThreshold  uint64 `json:"tune_threshold"`
}

type ModelTrainer struct {
	viewModelRepository ViewModelRepository
	transactor          postgresql.Transactor

	producer        Producer
	goCronScheduler *gocron.Scheduler
	thresholds      map[string]ModelTrainThreshold

	logger *zap.Logger
}

func NewModelTrainer(
	viewModelRepository ViewModelRepository,
	transactor postgresql.Transactor,
	producer Producer,
	goCronScheduler *gocron.Scheduler,
	thresholds map[string]ModelTrainThreshold,
	logger *zap.Logger,
) *ModelTrainer {
	return &ModelTrainer{
		viewModelRepository: viewModelRepository,
		transactor:          transactor,
		producer:            producer,
		goCronScheduler:     goCronScheduler,
		thresholds:          thresholds,
		logger:              logger,
	}
}

func (m ModelTrainer) StartTrainModels(cron string) {
	op := "model_trainer.ModelTrainer.StartTrainModels"
	_, err := m.goCronScheduler.CronWithSeconds(cron).Do(m.trainAndTuneModels)
	if err != nil {
		m.logger.Fatal(fmt.Sprintf("%s: %s", op, err.Error()))
	}
	m.goCronScheduler.StartBlocking()
}

func (m ModelTrainer) trainAndTuneModels() {
	op := "model_trainer.ModelTrainer.trainAndTuneModels"

	ctx := logger.ContextWithLogger(context.Background(), m.logger)

	txErr := m.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {

		for modelType, threshold := range m.thresholds {
			models, err := m.viewModelRepository.ViewNotLearnedModels(txCtx, modelType, threshold.TrainThreshold)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			for _, model := range models {
				err := m.viewModelRepository.SetModelStatus(txCtx, data.StatusInTrainProcess, modelType, model.UserID)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				msg, err := json.Marshal(map[string]string{"user_id": model.UserID, "model_type": modelType})
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				err = m.producer.Publish(modelType, msg)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			}

			models, err = m.viewModelRepository.ViewNotFineTunedFaceModels(txCtx, modelType, threshold.TuneThreshold)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			for _, model := range models {
				err := m.viewModelRepository.SetModelStatus(txCtx, data.StatusInTuneProcess, modelType, model.UserID)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				msg, err := json.Marshal(map[string]string{"user_id": model.UserID, "model_type": modelType})
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				err = m.producer.Publish(modelType, msg)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			}

			m.logger.Info(fmt.Sprintf("%s: train models worker finished", op))
		}

		return nil
	})

	if txErr != nil {
		m.logger.Error(txErr.Error())
	}
}
