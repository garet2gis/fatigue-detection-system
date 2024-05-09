package model_trainer

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/domains/data"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/workers"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/rabbitmq"
	"github.com/go-co-op/gocron"
	"github.com/urfave/cli/v2"
	"time"
)

func Action(_ *cli.Context) error {
	cfg := config.GetConfigModelTrainer()

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

	var queues []string
	for queue := range cfg.ModelTrainThresholds {
		queues = append(queues, queue)
	}

	err = rabbit.InitQueues(queues)
	if err != nil {
		l.Fatal(err.Error())
	}

	scheduler := gocron.NewScheduler(time.UTC)
	trainer := workers.NewModelTrainer(data.NewRepository(dbClient), dbClient, rabbit, scheduler, cfg.ModelTrainThresholds, l)
	if err != nil {
		l.Fatal(err.Error())
	}

	trainer.StartTrainModels(cfg.CRON)

	return nil
}
