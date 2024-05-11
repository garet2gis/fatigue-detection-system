package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/face_features_storage/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/face_features_storage/pkg/logger"
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
	userID := r.FormValue("user_id")
	modelType := r.FormValue("model_type")

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}(file)

	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		var featuresCount uint64
		featuresCount, err = c.dataRepository.SaveFaceVideoFeatures(txCtx, file)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.sendFeaturesCount(userID, modelType, int(featuresCount))
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

type IncreaseFeaturesRequest struct {
	ModelType     string `json:"model_type"  validate:"required"`
	UserID        string `json:"user_id"  validate:"required"`
	FeaturesCount int    `json:"features_count"  validate:"required"`
}

func (c *CoreHandler) sendFeaturesCount(userID, modelType string, featuresCount int) error {
	op := "handlers.CoreHandler.sendFeaturesCount"

	req := IncreaseFeaturesRequest{
		ModelType:     modelType,
		UserID:        userID,
		FeaturesCount: featuresCount,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	r, err := http.NewRequest("POST", c.storageHandlerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	r.Header.Set("Content-Type", "application/json")

	// Создаем клиент HTTP для отправки запроса
	client := &http.Client{}

	// Отправляем запрос
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s: error from model_handler service", op)
	}

	return nil
}
