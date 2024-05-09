package handlers

import (
	"context"
	"errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type FeatureRepository interface {
	ChangeFeaturesCount(ctx context.Context, userID, modelType string, faceFeaturesCount int) error
	CreateModel(ctx context.Context, userID, modelType string) error
	GetModelByUserID(ctx context.Context, userID, modelType string) (*data.MLModel, error)
	GetModelsByUserID(ctx context.Context, userID string) ([]data.MLModel, error)
	SetFeaturesCountUsed(ctx context.Context, userID, modelType string, faceFeaturesCount int) error
	SetModelS3Key(ctx context.Context, s3Key, modelType, userID string) error
	SetModelStatus(ctx context.Context, status string, modelType string, userID string) error
}

type ModelSaver interface {
	SaveFile(ctx context.Context, fileName string, file io.Reader) error
	GetPresignURL(ctx context.Context, fileName string) (string, error)
}

type Producer interface {
	Publish(queue string, message []byte) error
}

type CoreHandler struct {
	featureRepository FeatureRepository
	modelSaver        ModelSaver
	resultQueue       string
	transactor        postgresql.Transactor
	validator         *validator.Validate

	logger *zap.Logger
}

func NewCoreHandler(modelSaver ModelSaver,
	featureRepository FeatureRepository,
	transactor postgresql.Transactor,
	validator *validator.Validate,
	logger *zap.Logger) *CoreHandler {
	return &CoreHandler{
		featureRepository: featureRepository,
		transactor:        transactor,
		modelSaver:        modelSaver,
		validator:         validator,
		logger:            logger,
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

	router.Route("/api/v1", func(router chi.Router) {
		router.Post("/save_model", ErrorMiddleware(c.SaveModel))
		router.Post("/increase_features", ErrorMiddleware(c.IncreaseFeatures))
		router.Post("/get_models", ErrorMiddleware(c.GetModels))
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		l := logger.EntryWithRequestIDFromContext(r.Context())

		api.WriteError(r.Context(), w, app_errors.ErrNotFound.SetMessage("not found page").ToCoreError(), l)
	})

	return router
}
