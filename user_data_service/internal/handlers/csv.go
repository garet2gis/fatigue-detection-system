package handlers

import (
	"context"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"mime/multipart"
	"net/http"
)

// SaveCSV godoc
//
//	@Summary	Принимает csv файл и сохраняет информацию в БД
//	@ID			save csv
//	@Tags		Save CSV
//	@Param		file	formData	file	true	"Загружаемый csv"
//	@Success	204
//	@Failure	400	{object}	app_errors.AppError
//	@Router		/save_csv [post]
func (c *CoreHandler) SaveCSV(w http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.SaveCSV"

	l := logger.EntryWithRequestIDFromContext(r.Context())

	file, _, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			l.Error(fmt.Errorf("%s: %w", op, err).Error())
		}
	}(file)

	txErr := c.transactor.WithinTransaction(r.Context(), func(txCtx context.Context) error {
		err = c.dataRepository.CopyCSV(txCtx, file)
		if err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		return txErr
	}

	api.WriteSuccess(r.Context(), w, struct{}{}, http.StatusNoContent, l)
	return nil
}
