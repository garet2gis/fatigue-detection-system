package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"strconv"
)

type ViewModelRepository interface {
	ViewNotLearnedModels(ctx context.Context, modelType string, trainThreshold uint64) ([]data.MLModel, error)
	ViewNotFineTunedFaceModels(ctx context.Context, modelType string, tuneThreshold uint64) ([]data.MLModel, error)
	SetModelStatus(ctx context.Context, status string, modelType string, userID string) error
}

type Producer interface {
	Publish(queue string, message []byte) error
}

type Presigner interface {
	GetPresignURL(ctx context.Context, fileName string) (string, error)
}

type ModelTrainThreshold struct {
	TrainThreshold uint64 `json:"train_threshold"`
	TuneThreshold  uint64 `json:"tune_threshold"`
}

type ModelTrainer struct {
	viewModelRepository ViewModelRepository
	presigner           Presigner
	transactor          postgresql.Transactor

	producer        Producer
	goCronScheduler *gocron.Scheduler
	thresholds      map[string]ModelTrainThreshold

	logger *zap.Logger
}

func NewModelTrainer(
	viewModelRepository ViewModelRepository,
	presigner Presigner,
	transactor postgresql.Transactor,
	producer Producer,
	goCronScheduler *gocron.Scheduler,
	thresholds map[string]ModelTrainThreshold,
	logger *zap.Logger,
) *ModelTrainer {
	return &ModelTrainer{
		presigner:           presigner,
		viewModelRepository: viewModelRepository,
		transactor:          transactor,
		producer:            producer,
		goCronScheduler:     goCronScheduler,
		thresholds:          thresholds,
		logger:              logger,
	}
}

// StartTrainModels - функция запуска инициализации тренировки моделей по расписанию
func (m ModelTrainer) StartTrainModels(cron string) {
	// Объявляем текущую операцию для оборачивания ошибки
	op := "model_trainer.ModelTrainer.StartTrainModels"
	// Формируем задачу по расписанию в формате cron
	_, err := m.goCronScheduler.CronWithSeconds(cron).Do(m.trainAndTuneModels)
	if err != nil {
		m.logger.Fatal(fmt.Sprintf("%s: %s", op, err.Error()))
	}
	// Запускаем задачу синхронным методом
	m.goCronScheduler.StartBlocking()
}

// trainAndTuneModels - функция, отправляющая задачи на обучение и дообучение моделей
func (m ModelTrainer) trainAndTuneModels() {
	// Объявляем текущую операцию для оборачивания ошибки
	op := "model_trainer.ModelTrainer.trainAndTuneModels"

	// Кладем логгер в контекст
	ctx := logger.ContextWithLogger(context.Background(), m.logger)

	// Производим все операции изменения данных в транзакции
	txErr := m.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {

		// Перебираем все пороговые значение для моделей разных типов
		for modelType, threshold := range m.thresholds {
			// Находим модели определенного типа, количество признаков, у которых преодолел порог обучения
			models, err := m.viewModelRepository.ViewNotLearnedModels(txCtx, modelType, threshold.TrainThreshold)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			// Проходим по всем найденным моделям
			for _, model := range models {
				// Задаем статус модели - в процессе обучения
				err := m.viewModelRepository.SetModelStatus(txCtx, data.StatusInTrainProcess, modelType, model.UserID)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				// Формируем задачу на обучение
				msg, err := json.Marshal(map[string]string{"type": "train", "user_id": model.UserID, "model_type": modelType})
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				// Отправляем задачу в соответствующую очередь
				err = m.producer.Publish(modelType, msg)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			}

			// Находим модели определенного типа, количество признаков, у которых преодолел порог дообучения
			models, err = m.viewModelRepository.ViewNotFineTunedFaceModels(txCtx, modelType, threshold.TuneThreshold)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			// Проходим по всем найденным моделям
			for _, model := range models {
				// Задаем статус модели - в процессе дообучения
				err := m.viewModelRepository.SetModelStatus(txCtx, data.StatusInTuneProcess, modelType, model.UserID)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				// Формируем ссылку на скачивание предыдущей модели
				modelURL, err := m.presigner.GetPresignURL(txCtx, *model.S3Key)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				// Формируем задачу на дообучение
				msg, err := json.Marshal(map[string]string{
					"type":           "tune",
					"user_id":        model.UserID,
					"model_type":     modelType,
					"model_features": strconv.FormatUint(model.ModelFeatures, 10),
					"model_url":      modelURL})
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}

				// Отправляем задачу в соответствующую очередь
				err = m.producer.Publish(modelType, msg)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			}

			// Логгируем успешное выполнение воркера
			m.logger.Info(fmt.Sprintf("%s: train models worker finished", op))
		}

		return nil
	})

	if txErr != nil {
		m.logger.Error(txErr.Error())
	}
}
