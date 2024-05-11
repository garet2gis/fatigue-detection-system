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
	op := "handlers.CoreHandler.IncreaseFeatures"
	l := logger.EntryWithRequestIDFromContext(r.Context())

	var req fixtures.IncreaseFeaturesRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		_, err = c.featureRepository.GetModelByUserID(txCtx, req.UserID, data.FaceModel)
		if err != nil {
			if app_errors.IsNotFound(err) {
				err = c.featureRepository.CreateModel(txCtx, req.UserID, data.FaceModel)
				if err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
			} else {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		err = c.featureRepository.ChangeFeaturesCount(txCtx, req.UserID, data.FaceModel, req.FeaturesCount)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})

	if txErr != nil {
		return txErr
	}

	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}
