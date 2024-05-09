package main

import (
	"context"
	_ "github.com/garet2gis/fatigue-detection-system/model_handler_service/docs"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/handlers"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/s3_client"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/server"
)

//	@title		Model storage service
//	@version	1.0

//	@BasePath	/api/v1/

func main() {
	cfg := config.GetConfig()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	s3Client, err := s3_client.NewS3Client(context.Background(), cfg.ToS3Config())
	if err != nil {
		l.Fatal(err.Error())
	}

	coreHandler := handlers.NewCoreHandler(s3Client, l)

	app := server.NewServer(cfg.ToAppConfig(), coreHandler.Router(), l)

	app.Start()
}
