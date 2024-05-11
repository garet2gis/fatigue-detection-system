package config

import (
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/workers"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"sync"
)

type ModelTrainerConfig struct {
	DBConfig
	LoggerConfig

	RabbitURL      string `env:"RABBIT_URL_MT"  env-required:"true"`
	RabbitPoolSize int    `env:"RABBIT_POOL_SIZE_MT"  env-default:"10"`
	CRON           string `env:"CRON_MT"  env-default:"*/5 * * * * *"`

	PathToTrainThresholds string `env:"PATH_TO_TRAIN_THRESHOLDS"  env-default:"thresholds.json"`
	ModelTrainThresholds  map[string]workers.ModelTrainThreshold
}

func (c ModelTrainerConfig) ToDBConfig() postgresql.DBConfig {
	return postgresql.DBConfig{
		Port:                  c.DBPort,
		Host:                  c.DBHost,
		Name:                  c.DBName,
		Password:              c.DBPassword,
		Username:              c.DBUsername,
		MaxConnectionAttempts: c.MaxConnectionAttempts,
		AutoMigrate:           c.AutoMigrate,
		MigrationsDir:         c.MigrationsDir,
	}
}

func (c ModelTrainerConfig) ToLoggerConfig() logger.LoggerConfig {
	return logger.LoggerConfig{
		IsProduction: c.IsProduction,
	}
}

var instanceModelTrainer *ModelTrainerConfig
var onceModelTrainer sync.Once

func GetConfigModelTrainer() *ModelTrainerConfig {
	onceModelTrainer.Do(func() {
		log.Print("Read application configuration")

		instanceModelTrainer = &ModelTrainerConfig{}
		if err := cleanenv.ReadEnv(instanceModelTrainer); err != nil {
			help, _ := cleanenv.GetDescription(instanceModelTrainer, nil)

			log.Print(help)
			log.Fatal(err)
		}
		var err error
		instanceModelTrainer.ModelTrainThresholds, err = parseTrainThresholdConfig(instanceModelTrainer.PathToTrainThresholds)
		if err != nil {
			log.Fatal(err)
		}
	})

	return instanceModelTrainer
}

func parseTrainThresholdConfig(path string) (map[string]workers.ModelTrainThreshold, error) {
	op := "config.parseTrainThresholdConfig"
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("%s: failed to close file due to err: %s\n", op, err)
		}
	}(file)

	dec := json.NewDecoder(file)

	trainThresholds := make(map[string]workers.ModelTrainThreshold)
	err = dec.Decode(&trainThresholds)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return trainThresholds, nil
}
