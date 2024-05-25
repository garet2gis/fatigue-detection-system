package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/handlers/fixtures"
	customTools "github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/tools"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"

	"net/http"
)

// IncreaseFeatures godoc
//
//	@Summary	Принимает количество новых фич по моделям
//	@ID			increase features
//	@Tags		Features
//	@Param		features_data	body	fixtures.IncreaseFeaturesRequest	true	"Данные о количестве фич"
//	@Success	204
//	@Failure	400	{object}	app_errors.AppError
//	@Router		/increase_features [post]
func (c *CoreHandler) IncreaseFeatures(w http.ResponseWriter, r *http.Request) error {
	// Объявляем текущую операцию для оборачивания ошибки
	op := "handlers.CoreHandler.IncreaseFeatures"
	// Берем логгер из контекста
	l := logger.EntryWithRequestIDFromContext(r.Context())

	// Десереализуем данные из тела запроса
	var req fixtures.IncreaseFeaturesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	// Валидируем данные на наличие необходимых полей
	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	// Делаем все изменения данных в БД в транзакции
	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		// Берем модель пользователя по ID пользователя и типу модели из БД
		_, err = c.featureRepository.GetModelByUserID(txCtx, req.UserID, data.FaceModel)
		if err != nil {
			// Если модель не нашлась, то создаем новоую в БД
			if app_errors.IsNotFound(err) {
				err = c.featureRepository.CreateModel(txCtx, req.UserID, data.FaceModel)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			} else {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		// увеличиваем количество признаков на желаемое значение в БД
		err = c.featureRepository.ChangeFeaturesCount(txCtx, req.UserID, data.FaceModel, req.FeaturesCount)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})

	if txErr != nil {
		return txErr
	}

	// Возвращаем пустой ответ со статусом 204
	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}
