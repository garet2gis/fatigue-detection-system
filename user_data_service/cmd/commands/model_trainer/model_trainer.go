package model_trainer

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/workers"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/rabbitmq"
	"github.com/go-co-op/gocron"
	"github.com/urfave/cli/v2"
	"time"
)

func Action(_ *cli.Context) error {
	cfg := config.GetConfig()

	// db
	// logger
	// TODO: add to cfg
	rabbitURL := "amqp://user:password@localhost:5672/"
	poolSize := 10
	cfgThresholds := map[string]workers.ModelTrainThreshold{data.FaceModel: {TrainThreshold: 8000, TuneThreshold: 2000}}
	cron := "*/3 * * * * *"

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

	scheduler := gocron.NewScheduler(time.UTC)
	trainer := workers.NewModelTrainer(data.NewRepository(dbClient), dbClient, rabbit, scheduler, cfgThresholds, l)
	if err != nil {
		l.Fatal(err.Error())
	}

	// каждые 3 секунды
	trainer.StartTrainModels(cron)

	return nil
}
