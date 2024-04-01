package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

type AppConfig struct {
	Host string
	Port string

	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int

	GracefulShutdownTimeout uint

	IsEnableSwagger bool
	SwaggerURL      string
}

type Server struct {
	router chi.Router

	cfg              AppConfig
	logger           *zap.Logger
	shutdownCallback func(ctx context.Context) error
}

func NewServer(cfg AppConfig, router chi.Router, logger *zap.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		router: router,
	}
}

func (s *Server) SetShutdownCallback(shutdownCallback func(ctx context.Context) error) {
	s.shutdownCallback = shutdownCallback
}

func (s *Server) SetRouter(router chi.Router) {
	s.router = router
}

func (s *Server) Start() {
	s.initSwagger()
	s.startServerGracefully()
}

func (s *Server) initSwagger() {
	if !s.cfg.IsEnableSwagger {
		return
	}

	url := s.cfg.SwaggerURL

	if url == "" {
		url = fmt.Sprintf("http://%s:%s/swagger/doc.json", s.cfg.Host, s.cfg.Port)
	} else {
		url += "/swagger/doc.json"
	}

	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(url),
	))
}

// StartServerGracefully запускает сервер с параметрами конфига.
// timeout - время для завершения graceful shutdown
func (s *Server) startServerGracefully() {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.cfg.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(s.cfg.WriteTimeoutSeconds) * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, cancel := context.WithTimeout(serverCtx, time.Duration(s.cfg.GracefulShutdownTimeout)*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				s.logger.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			s.logger.Fatal(err.Error())
		}
		s.logger.Info("server shutdown gracefully")

		if s.shutdownCallback != nil {
			err = s.shutdownCallback(shutdownCtx)
			if err != nil {
				s.logger.Fatal(err.Error())
			}
			s.logger.Info("executed shutdown callback")
		}

		serverStopCtx()
	}()

	s.logger.With(zap.String("addr", addr)).Info("server started")
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal(err.Error())
	}

	<-serverCtx.Done()
}
