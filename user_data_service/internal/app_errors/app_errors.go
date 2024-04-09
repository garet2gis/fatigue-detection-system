package app_errors

import (
	"errors"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"net/http"
)

type AppError struct {
	// Наименование ошибки
	Name string `json:"name" validate:"required" example:"NotFound"`
	// Сообщение ошибки
	Message string `json:"message" validate:"required" example:"entity not found"`
	// Код ошибки
	Code int `json:"code" validate:"required" example:"26002"`
	// Статус код ответа
	Status int `json:"status" validate:"required" example:"404"`
	// Начальная ошибка
	InternalError error `json:"-"`
	// Нужно ли логировать ошибку в миддлваре
	IsNotLogging bool `json:"-"`
} //	@AppError

func (e AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.InternalError != nil {
		e.InternalError.Error()
	}

	return "undefined error"
}

func NewAppError(name, message string, code, status int) *AppError {
	return &AppError{
		Name:    name,
		Message: message,
		Code:    code,
		Status:  status,
	}
}

func (e AppError) SetMessage(message string) *AppError {
	e.Message = message
	return &e
}

func (e AppError) DisableLog() *AppError {
	e.IsNotLogging = true
	return &e
}

func (e AppError) WrapError(op, msg string) error {
	return fmt.Errorf("%s: %w", op, e.SetMessage(msg))
}

func (e AppError) SetError(error error) *AppError {
	e.InternalError = error
	return &e
}

func (e AppError) ToCoreError() api.AppError {
	return api.AppError{
		Name:    e.Name,
		Message: e.Message,
		Code:    e.Code,
		Status:  e.Status,
	}
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Status == http.StatusNotFound {
			return true
		}
	}
	return false
}
