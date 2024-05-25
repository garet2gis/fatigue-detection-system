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
	// Объявляем текущую операцию для оборачивания ошибки
	op := "handlers.CoreHandler.SaveModel"
	// Берем логгер из контекста
	l := logger.EntryWithRequestIDFromContext(r.Context())

	// Берем модель из формы по ключу file
	file, header, err := r.FormFile("file")
	if err != nil {
		return err
	}
	// Закрываем файл при завершении функции
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Errorf("%s: %w", op, err).Error())
		}
	}(file)
	// Берем строковое значение user_id из формы
	userID := r.FormValue("user_id")
	// Берем строковое значение model_type из формы
	modelType := r.FormValue("model_type")
	// Берем строковое значение features_count из формы
	featuresCountString := r.FormValue("features_count")

	// Преобразуем features_count к типу int
	featuresCount, err := strconv.Atoi(featuresCountString)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	// Создаем новый uuid для модели
	modelID, err := uuid.NewUUID()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Задаем путь и название файла в s3
	filename := path.Join(modelType, userID, fmt.Sprintf("%s_%s", modelID.String(), header.Filename))

	// Делаем все изменения данных в транзакции
	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		// Задаем новый s3 ключ для модели
		err := c.featureRepository.SetModelS3Key(txCtx, filename, modelType, userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		// Задаем статус модели - обучена
		err = c.featureRepository.SetModelStatus(txCtx, data.StatusTrained, modelType, userID)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		// Задаем количество признаков, которые использовались в обучении
		err = c.featureRepository.SetFeaturesCountUsed(txCtx, userID, modelType, featuresCount)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		// Сохраняем модель в s3
		err = c.modelSaver.SaveFile(r.Context(), filename, file)
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
	// Объявляем текущую операцию для оборачивания ошибки
	op := "handlers.CoreHandler.GetModels"
	// Берем логгер из контекста
	l := logger.EntryWithRequestIDFromContext(r.Context())

	// Десереализуем данные из тела запроса
	var req fixtures.GetModelsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	// Валидируем данные на наличие необходимых полей
	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	// Находим модели данного пользователя
	models, err := c.featureRepository.GetModelsByUserID(r.Context(), req.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Формируем ответ, состоящий из предподписанных ссылок на скачивание всех типов моделей
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

	// Возвращаем результат со статусом 200
	api.WriteSuccess(r.Context(), w, modelsURLS, http.StatusOK, l)
	return nil
}
