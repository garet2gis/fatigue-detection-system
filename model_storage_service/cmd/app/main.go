package main

import (
	"context"
	_ "github.com/garet2gis/fatigue-detection-system/model_storage_service/docs"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/internal/handlers"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/rabbitmq"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/s3_client"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/server"
)

//	@title		Model storage service
//	@version	1.0

//	@BasePath	/api/v1/

func main() {
	cfg := config.GetConfig()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	rabbit, err := rabbitmq.NewRabbitMQConnection(cfg.RabbitURL, cfg.RabbitPoolSize)
	if err != nil {
		l.Fatal(err.Error())
	}
	defer rabbit.Close()

	err = rabbit.InitQueues([]string{cfg.ResultQueue})
	if err != nil {
		l.Fatal(err.Error())
	}

	s3Client, err := s3_client.NewS3Client(context.Background(), cfg.ToS3Config())
	if err != nil {
		l.Fatal(err.Error())
	}

	coreHandler := handlers.NewCoreHandler(s3Client, rabbit, cfg.RabbitConfig.ResultQueue, l)

	app := server.NewServer(cfg.ToAppConfig(), coreHandler.Router(), l)

	app.SetShutdownCallback(func(_ context.Context) error {
		rabbit.Close()
		return nil
	})

	app.Start()
}
