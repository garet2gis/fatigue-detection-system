package app_errors

import "github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/api"

type AppError struct {
	// Наименование ошибки
	Name string `json:"name" validate:"required" example:"NotFound"`
	// Сообщение ошибки
	Message string `json:"message" validate:"required" example:"entity not found"`
	// Код ошибки
	Code int `json:"code" validate:"required" example:"26002"`
	// Статус код ответа
	Status int `json:"status" validate:"required" example:"404"`
} //	@AppError

func (e AppError) Error() string {
	return e.Message
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

func (e AppError) ToCoreError() api.AppError {
	return api.AppError{
		Name:    e.Name,
		Message: e.Message,
		Code:    e.Code,
		Status:  e.Status,
	}
}
