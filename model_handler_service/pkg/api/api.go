package api

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

type Success struct {
	Status    string      `json:"status"`
	Content   interface{} `json:"content"`
	RequestID string      `json:"request_id"`
}

func WriteSuccess(ctx context.Context, w http.ResponseWriter, success interface{}, status int, l *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	requestID, ok := ctx.Value(middleware.RequestIDHeader).(string)
	if !ok {
		requestID = "unknown"
	}

	response := Success{
		Status:    "success",
		Content:   success,
		RequestID: requestID,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		l.With(zap.Error(err)).Error("failed to marshal success json")
	}
	w.Write(jsonResponse)
}

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

func (e AppError) SetMessage(msg string) *AppError {
	e.Message = msg
	return &e
}

type Error struct {
	Status    string   `json:"status"`
	Error     AppError `json:"error"`
	RequestID string   `json:"request_id"`
}

func WriteError(ctx context.Context, w http.ResponseWriter, error AppError, l *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(error.Status)

	requestID, ok := ctx.Value(middleware.RequestIDHeader).(string)
	if !ok {
		requestID = "unknown"
	}

	response := Error{
		Status:    "error",
		Error:     error,
		RequestID: requestID,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		l.With(zap.Error(err)).Error("failed to marshal error json")
	}
	w.Write(jsonResponse)
}
