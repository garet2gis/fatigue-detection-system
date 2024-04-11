package config

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type ModelUpdaterConfig struct {
	DBConfig
	LoggerConfig

	RabbitURL      string `env:"RABBIT_URL_MU"  env-required:"true"`
	RabbitPoolSize int    `env:"RABBIT_POOL_SIZE_MU"  env-default:"1"`
	ResultQueue    string `env:"RESULT_QUEUE"  env-default:"result"`
}

func (c ModelUpdaterConfig) ToDBConfig() postgresql.DBConfig {
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

func (c ModelUpdaterConfig) ToLoggerConfig() logger.LoggerConfig {
	return logger.LoggerConfig{
		IsProduction: c.IsProduction,
	}
}

var instanceModelUpdater *ModelUpdaterConfig
var onceModelUpdater sync.Once

func GetConfigModelUpdater() *ModelUpdaterConfig {
	onceModelUpdater.Do(func() {
		log.Print("Read application configuration")

		instanceModelUpdater = &ModelUpdaterConfig{}
		if err := cleanenv.ReadEnv(instanceModelUpdater); err != nil {
			help, _ := cleanenv.GetDescription(instanceModelUpdater, nil)

			log.Print(help)
			log.Fatal(err)
		}
	})

	return instanceModelUpdater
}
