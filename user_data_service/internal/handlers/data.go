package handlers

import (
	"bytes"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"io"
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

	file, header, err := r.FormFile("file")
	if err != nil {
		return err
	}

	jwt := r.URL.Query().Get("access_token")
	if jwt == "" {
		return app_errors.ErrUnauthorized.WrapError(op, "empty access_token")
	}

	userID, err := c.tokenGenerator.GetUserIDFromToken(r.Context(), jwt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Sprintf("%s: %v", op, err))
		}
	}(file)

	err = c.sendFeatures(file, header.Filename, userID, "face_model")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}

func (c *CoreHandler) sendFeatures(file multipart.File, fileName, userID, modelType string) error {
	op := "handlers.CoreHandler.sendFeatures"
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = writer.WriteField("user_id", userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = writer.WriteField("model_type", modelType)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Закрываем writer, чтобы записать завершающую границу
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req, err := http.NewRequest("POST", c.FeaturesURL, &requestBody)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%s: error from video_features service", op)
	}

	return nil
}
