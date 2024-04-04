package handlers

import (
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/logger"

	"mime/multipart"
	"net/http"
)

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

	err = c.modelSaver.SaveFile(r.Context(), header.Filename, file)
	if err != nil {
		return err
	}

	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}
