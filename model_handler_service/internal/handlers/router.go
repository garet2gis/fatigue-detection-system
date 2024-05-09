package handlers

import (
	"context"
	"errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"io"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type ModelSaver interface {
	SaveFile(ctx context.Context, fileName string, file io.Reader) error
	GenerateS3DownloadLink(key string) (string, error)
}

type Producer interface {
	Publish(queue string, message []byte) error
}

type CoreHandler struct {
	modelSaver  ModelSaver
	resultQueue string

	logger *zap.Logger
}

func NewCoreHandler(modelSaver ModelSaver, logger *zap.Logger) *CoreHandler {
	return &CoreHandler{
		modelSaver: modelSaver,
		logger:     logger,
	}
}

func InitRequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(middleware.RequestIDHeader)
		if requestID == "" {
			id, _ := uuid.NewUUID()
			requestID = strings.ReplaceAll(id.String(), "-", "")
		}
		ctx = context.WithValue(ctx, middleware.RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

func ErrorMiddleware(handler ErrorHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			var appErr *app_errors.AppError

			// если неизвестная ошибка, то устанавливаем UnknownError
			// и переопределяем ошибку внутри (возможно понадобиться для логов)
			if !errors.As(err, &appErr) {
				appErr = app_errors.UnknownError.SetMessage(err.Error())
			}

			l := logger.EntryWithRequestIDFromContext(r.Context())

			api.WriteError(r.Context(), w, appErr.ToCoreError(), l)
		}
	}
}

func (c *CoreHandler) Router() chi.Router {
	router := chi.NewRouter()

	router.Use(InitRequestID)
	router.Use(logger.WithLogger(c.logger))

	router.Post("/api/v1/save_model", ErrorMiddleware(c.SaveModel))

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		l := logger.EntryWithRequestIDFromContext(r.Context())

		api.WriteError(r.Context(), w, app_errors.ErrNotFound.SetMessage("not found page").ToCoreError(), l)
	})

	return router
}
