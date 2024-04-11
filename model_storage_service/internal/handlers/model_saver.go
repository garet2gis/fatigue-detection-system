package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/logger"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
)

type Message struct {
	ModelType     string `json:"model_type"`
	UserID        string `json:"user_id"`
	ModelURL      string `json:"model_url"`
	FeaturesCount int    `json:"features_count"`
}

// SaveModel godoc
//
//	@Summary	Принимает ml модель
//	@ID			save model
//	@Tags		Save model
//	@Param		file	formData	file	true	"Загружаемая ml-модель"
//	@Success	204
//	@Failure	400	{object}	app_errors.AppError
//	@Router		/save_model [post]
func (c *CoreHandler) SaveModel(w http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.SaveModel"
	l := logger.EntryWithRequestIDFromContext(r.Context())

	file, header, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Errorf("%s: %w", op, err).Error())
		}
	}(file)

	userID := r.FormValue("user_id")
	modelType := r.FormValue("model_type")
	featuresCountString := r.FormValue("features_count")

	featuresCount, err := strconv.Atoi(featuresCountString)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	modelID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	filename := path.Join(modelType, userID, fmt.Sprintf("%s_%s", modelID.String(), header.Filename))

	err = c.modelSaver.SaveFile(r.Context(), filename, file)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	url, err := c.modelSaver.GenerateS3DownloadLink(filename)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	resultMessage := Message{
		ModelType:     modelType,
		UserID:        userID,
		ModelURL:      url,
		FeaturesCount: featuresCount,
	}

	msg, err := json.Marshal(resultMessage)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = c.producer.Publish(c.resultQueue, msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}
