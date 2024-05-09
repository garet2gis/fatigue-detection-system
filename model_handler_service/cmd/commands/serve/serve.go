package serve

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/handlers"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/s3_client"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/server"
	"github.com/go-playground/validator/v10"
	"github.com/urfave/cli/v2"
)

func Action(_ *cli.Context) error {
	cfg := config.GetConfig()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	s3Client, err := s3_client.NewS3Client(context.Background(), cfg.ToS3Config())
	if err != nil {
		l.Fatal(err.Error())
	}

	dbClient, err := postgresql.NewClient(context.Background(), cfg.ToDBConfig())
	if err != nil {
		l.Fatal(err.Error())
	}
	validate := validator.New()

	coreHandler := handlers.NewCoreHandler(s3Client, data.NewRepository(dbClient), dbClient, validate, l)

	app := server.NewServer(cfg.ToAppConfig(), coreHandler.Router(), l)

	app.SetShutdownCallback(func(_ context.Context) error {
		dbClient.Close()
		return nil
	})

	app.Start()

	return nil
}
