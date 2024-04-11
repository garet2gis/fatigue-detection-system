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
	cfg := config.GetConfig()

	//db
	//logger
	// TODO: add to cfg
	rabbitURL := "amqp://user:password@localhost:5672/"
	poolSize := 1
	resQueue := "result"

	l := logger.NewLogger(cfg.ToLoggerConfig())

	dbClient, err := postgresql.NewClient(context.Background(), cfg.ToDBConfig())
	if err != nil {
		l.Fatal(err.Error())
	}

	rabbit, err := rabbitmq.NewRabbitMQConnection(rabbitURL, poolSize)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rabbit.Close()

	updater := workers.NewModelUpdater(data.NewRepository(dbClient), dbClient, rabbit, resQueue, l)
	if err != nil {
		l.Fatal(err.Error())
	}

	updater.StartModelUpdate()

	return nil
}
