package handlers

import (
	"context"
	"errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/auth"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"strings"
)

type DataRepository interface {
	SaveFaceVideoFeatures(ctx context.Context, file multipart.File) (uint64, error)
	IncrementFeaturesCount(ctx context.Context, userID string, faceFeaturesCount uint64) error
	CreateFeaturesCount(ctx context.Context, userID string) error
	GetFeaturesCount(ctx context.Context, userID string) (*data.FeatureCount, error)
}

type AuthRepository interface {
	CreateUser(ctx context.Context, model auth.User) (string, error)
	GetUserByLogin(ctx context.Context, login string) (*auth.User, error)
}

type TokenGenerator interface {
	CreateAccessToken(_ context.Context, userID string) (string, error)
	GetUserIDFromToken(_ context.Context, contentToken string) (string, error)
}

type CoreHandler struct {
	dataRepository DataRepository
	authRepository AuthRepository
	tokenGenerator TokenGenerator
	transactor     postgresql.Transactor
	validator      *validator.Validate

	BaseURL string
	logger  *zap.Logger
}

func NewCoreHandler(
	dataRepository DataRepository,
	authRepository AuthRepository,
	tokenGenerator TokenGenerator,
	BaseURL string,
	transactor postgresql.Transactor,
	validator *validator.Validate,
	logger *zap.Logger,
) *CoreHandler {
	return &CoreHandler{
		validator:      validator,
		tokenGenerator: tokenGenerator,
		BaseURL:        BaseURL,
		dataRepository: dataRepository,
		authRepository: authRepository,
		transactor:     transactor,
		logger:         logger,
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

			if !appErr.IsNotLogging {
				// увеличиваю уровень логирования stacktrace до panic,
				// чтобы в логе не было бесполезного stacktrace
				l.WithOptions(zap.AddStacktrace(zap.DPanicLevel)).Error(err.Error())
			}

			api.WriteError(r.Context(), w, appErr.ToCoreError(), l)
		}
	}
}

func (c *CoreHandler) Router() chi.Router {
	router := chi.NewRouter()

	router.Use(InitRequestID)
	router.Use(logger.WithLogger(c.logger))

	router.Route("/api/v1", func(router chi.Router) {
		router.Route("/face_model", func(router chi.Router) {
			router.Post("/save_features", ErrorMiddleware(c.SaveVideoFeatures))
		})

		router.Route("/auth", func(router chi.Router) {
			router.Post("/register", ErrorMiddleware(c.Register))
			router.Post("/login", ErrorMiddleware(c.Login))
		})
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		l := logger.EntryWithRequestIDFromContext(r.Context())

		api.WriteError(r.Context(), w, app_errors.ErrNotFound.SetMessage("not found page").ToCoreError(), l)
	})

	return router
}
