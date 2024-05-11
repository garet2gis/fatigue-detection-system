package logger

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

type ctxLogger struct{}

// ContextWithLogger adds logger to context
func ContextWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

// LoggerFromContext returns logger from context
func LoggerFromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxLogger{}).(*zap.Logger); ok {
		return l
	}
	return zap.L()
}

// EntryWithRequestIDFromContext returns logger from context with request ID
func EntryWithRequestIDFromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxLogger{}).(*zap.Logger); ok {
		requestID, ok := ctx.Value(middleware.RequestIDHeader).(string)
		if ok {
			return l.With(zap.String("request_id", requestID))
		}

		return l.With(zap.String("request_id", "unknown"))
	}
	return zap.L().With(zap.String("request_id", "unknown"))
}

// WithLogger middleware для обогащение контекста запроса логером
func WithLogger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ctxLogger{}, l))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
