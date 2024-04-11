package model_updater

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/workers"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/rabbitmq"
	"github.com/urfave/cli/v2"
)

func Action(_ *cli.Context) error {
	cfg := config.GetConfigModelUpdater()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	dbClient, err := postgresql.NewClient(context.Background(), cfg.ToDBConfig())
	if err != nil {
		l.Fatal(err.Error())
	}

	rabbit, err := rabbitmq.NewRabbitMQConnection(cfg.RabbitURL, cfg.RabbitPoolSize)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rabbit.Close()

	updater := workers.NewModelUpdater(data.NewRepository(dbClient), dbClient, rabbit, cfg.ResultQueue, l)
	if err != nil {
		l.Fatal(err.Error())
	}

	updater.StartModelUpdate()

	return nil
}
