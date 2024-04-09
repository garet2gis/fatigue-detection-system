package main

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/auth"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/handlers"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/server"
	"github.com/go-playground/validator/v10"

	_ "github.com/garet2gis/fatigue-detection-system/user_data_service/docs"
	_ "github.com/garet2gis/fatigue-detection-system/user_data_service/migrations"
)

//	@title		User data service API
//	@version	1.0

//	@BasePath	/api/v1/

func main() {
	cfg := config.GetConfig()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	dbClient, err := postgresql.NewClient(context.Background(), cfg.ToDBConfig())
	if err != nil {
		l.Fatal(err.Error())
	}
	validate := validator.New()

	coreHandler := handlers.NewCoreHandler(
		data.NewRepository(dbClient),
		auth.NewRepository(dbClient),
		handlers.NewTokenHandler(cfg.JWTSecret),
		cfg.BaseURL,
		dbClient,
		validate,
		l)

	app := server.NewServer(cfg.ToAppConfig(), coreHandler.Router(), l)

	app.SetShutdownCallback(func(_ context.Context) error {
		dbClient.Close()
		return nil
	})

	app.Start()
}
