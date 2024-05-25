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
	// объявляем текущую операцию для оборачивания ошибки
	op := "handlers.CoreHandler.SaveVideoFeatures"
	// берем логгер из контекста
	l := logger.EntryWithRequestIDFromContext(r.Context())

	// берем файл из формы по ключу file
	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}
	// берем строковое значение user_id из формы
	userID := r.FormValue("user_id")
	// берем строковое значение model_type из формы
	modelType := r.FormValue("model_type")

	// закрываем файл при выходе из функции
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}(file)

	// делаем сохранение данных в транзакции
	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		var featuresCount uint64
		// сохраняем данные файла в таблице признаков
		featuresCount, err = c.dataRepository.SaveFaceVideoFeatures(txCtx, file)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		// отправляем количество сохраненных признаков в сервис работы с моделями
		err = c.sendFeaturesCount(userID, modelType, int(featuresCount))
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

type IncreaseFeaturesRequest struct {
	ModelType     string `json:"model_type"  validate:"required"`
	UserID        string `json:"user_id"  validate:"required"`
	FeaturesCount int    `json:"features_count"  validate:"required"`
}

func (c *CoreHandler) sendFeaturesCount(userID, modelType string, featuresCount int) error {
	// объявляем текущую операцию для оборачивания ошибки
	op := "handlers.CoreHandler.sendFeaturesCount"

	// определяем тело запроса
	req := IncreaseFeaturesRequest{
		ModelType:     modelType,
		UserID:        userID,
		FeaturesCount: featuresCount,
	}

	// сериализуем тело запроса в JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// создаем запрос по соответствующему url и кладем в него тело запроса
	r, err := http.NewRequest("POST", c.storageHandlerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// задаем тип контента равным application/json
	r.Header.Set("Content-Type", "application/json")

	// cоздаем клиент HTTP для отправки запроса
	client := &http.Client{}

	// отправляем запрос
	resp, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	// закрываем тело ответа при выходе из функции
	defer resp.Body.Close()

	// в случае ошибочного статуса ответа возвращаем ошибку
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s: error from model_handler service", op)
	}

	return nil
}
