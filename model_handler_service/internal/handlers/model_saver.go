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
//	@Tags		Models
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

	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		err := c.featureRepository.SetModelS3Key(txCtx, filename, modelType, userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.featureRepository.SetModelStatus(txCtx, data.StatusTrained, modelType, userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.featureRepository.SetFeaturesCountUsed(txCtx, userID, modelType, featuresCount)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		err = c.modelSaver.SaveFile(r.Context(), filename, file)
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

// GetModels godoc
//
//	@Summary	Возвращает ссылки на модели по id пользователя
//	@ID			get models
//	@Tags		Models
//	@Param		features_data	body		fixtures.GetModelsRequest	true	"ID пользователя"
//	@Success	200				{object}	map[string]string
//	@Failure	400				{object}	app_errors.AppError
//	@Router		/get_models [post]
func (c *CoreHandler) GetModels(w http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.GetModels"
	l := logger.EntryWithRequestIDFromContext(r.Context())

	var req fixtures.GetModelsRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	models, err := c.featureRepository.GetModelsByUserID(r.Context(), req.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	modelsURLS := make(map[string]string, len(models))
	for _, model := range models {
		if model.S3Key != nil {
			var modelURL string
			modelURL, err = c.modelSaver.GetPresignURL(r.Context(), *model.S3Key)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
			modelsURLS[model.ModelType] = modelURL
		}
	}

	api.WriteSuccess(r.Context(), w, modelsURLS, http.StatusOK, l)
	return nil
}
