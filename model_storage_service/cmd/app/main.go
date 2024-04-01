package main

import (
	"context"
	"fmt"
	_ "github.com/garet2gis/fatigue-detection-system/model_storage_service/docs"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/internal/config"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_storage_service/pkg/s3_client"
	"os"
)

//	@title		Model storage service
//	@version	1.0

//	@BasePath	/api/v1/

func main() {
	cfg := config.GetConfig()

	l := logger.NewLogger(cfg.ToLoggerConfig())

	//dbClient, err := postgresql.NewClient(context.Background(), cfg.ToDBConfig())
	//if err != nil {
	//	l.Fatal(err.Error())
	//}

	s3Client, err := s3_client.NewS3Client(context.Background(), cfg.ToS3Config())
	if err != nil {
		l.Fatal(err.Error())
	}

	fileToUpload := "./lol.jpeg"
	file, err := os.Open(fileToUpload)
	if err != nil {
		fmt.Println("failed to open file, ", err)
		return
	}
	defer file.Close()

	err = s3Client.UploadFile(context.Background(), "test_file.jpg", file)
	if err != nil {
		l.Fatal(err.Error())
	}
}
