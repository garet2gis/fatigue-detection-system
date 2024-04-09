package handlers

import (
	"context"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"mime/multipart"
	"net/http"
)

// SaveVideoFeatures godoc
//
//	@Summary	Принимает csv файл с фичами из видео
//	@ID			save csv
//	@Tags		Save CSV
//	@Param		file	formData	file	true	"Загружаемый csv"
//	@Success	204
//	@Failure	400	{object}	app_errors.AppError
//	@Router		/face_model/save_features [post]
func (c *CoreHandler) SaveVideoFeatures(w http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.SaveVideoFeatures"

	l := logger.EntryWithRequestIDFromContext(r.Context())

	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}

	// TODO: get user_id from access_token
	userID := "mock_user_id"

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}(file)

	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		featuresCount, err := c.dataRepository.SaveFaceVideoFeatures(txCtx, file)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.dataRepository.IncrementFeaturesCount(txCtx, userID, featuresCount)
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
