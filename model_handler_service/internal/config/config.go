package config

import (
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/s3_client"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/server"
	"log"

	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type DBConfig struct {
	DBPort                string `env:"DB_PORT" env-required:"true"`
	DBHost                string `env:"DB_HOST" env-required:"true"`
	DBName                string `env:"DB_NAME" env-required:"true"`
	DBPassword            string `env:"DB_PASSWORD" env-required:"true"`
	DBUsername            string `env:"DB_USERNAME" env-required:"true"`
	MaxConnectionAttempts int    `env:"MAX_ATTEMPTS" env-default:"10"`

	AutoMigrate   bool   `env:"AUTO_MIGRATE" env-default:"true"`
	MigrationsDir string `env:"MIGRATIONS_DIR" env-default:"migrations"`
}

type S3Config struct {
	Region string `env:"S3_REGION" env-default:"us-east-1"`
	S3Host string `env:"S3_HOST" env-default:"http://localhost:9000"`

	PartitionID       string `env:"PARTITION_ID" env-default:"aws"`
	HostnameImmutable bool   `env:"HOSTNAME_IMMUTABLE" env-default:"true"`
	BucketName        string `env:"BUCKET_NAME" env-default:"test"`

	AccessKeyID     string `env:"ACCESS_KEY_ID" env-required:"true"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY" env-required:"true"`
}

type LoggerConfig struct {
	IsProduction bool `env:"IS_PRODUCTION" env-default:"true"`
}

type HTTPConfig struct {
	Port                    string `env:"PORT"  env-required:"true"`
	Host                    string `env:"HOST"  env-required:"true"`
	ReadTimeoutSeconds      int    `env:"READ_TIMEOUT_SEC" env-default:"0"`
	WriteTimeoutSeconds     int    `env:"WRITE_TIMEOUT_SEC" env-default:"0"`
	GracefulShutdownTimeout uint   `env:"GRACEFUL_SHUTDOWN_TIMEOUT_SEC" env-default:"5"`
}

type SwaggerConfig struct {
	IsEnableSwagger bool   `env:"IS_ENABLE_SWAGGER" env-default:"true"`
	SwaggerURL      string `env:"SWAGGER_URL"`
}

type Config struct {
	DBConfig
	LoggerConfig
	HTTPConfig
	S3Config
	SwaggerConfig
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		log.Print("Read application configuration")

		instance = &Config{}
		if err := cleanenv.ReadEnv(instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)

			log.Print(help)
			log.Fatal(err)
		}
	})

	return instance
}

func (c Config) ToDBConfig() postgresql.DBConfig {
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

func (c Config) ToLoggerConfig() logger.LoggerConfig {
	return logger.LoggerConfig{
		IsProduction: c.IsProduction,
	}
}

func (c Config) ToAppConfig() server.AppConfig {
	return server.AppConfig{
		Host:                    c.Host,
		Port:                    c.Port,
		ReadTimeoutSeconds:      c.ReadTimeoutSeconds,
		WriteTimeoutSeconds:     c.WriteTimeoutSeconds,
		GracefulShutdownTimeout: c.GracefulShutdownTimeout,
		IsEnableSwagger:         c.IsEnableSwagger,
		SwaggerURL:              c.SwaggerURL,
	}
}

func (c Config) ToS3Config() s3_client.ConfigS3 {
	return s3_client.ConfigS3{
		Region:            c.S3Config.Region,
		S3Host:            c.S3Config.S3Host,
		PartitionID:       c.S3Config.PartitionID,
		HostnameImmutable: c.S3Config.HostnameImmutable,
		Bucket:            c.S3Config.BucketName,
		AccessKeyID:       c.S3Config.AccessKeyID,
		SecretAccessKey:   c.S3Config.SecretAccessKey,
	}
}
