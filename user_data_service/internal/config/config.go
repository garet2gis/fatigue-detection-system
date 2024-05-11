package config

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/server"
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

type HTTPConfig struct {
	Port                    string `env:"PORT"  env-required:"true"`
	Host                    string `env:"HOST"  env-required:"true"`
	ReadTimeoutSeconds      int    `env:"READ_TIMEOUT_SEC" env-default:"0"`
	WriteTimeoutSeconds     int    `env:"WRITE_TIMEOUT_SEC" env-default:"0"`
	GracefulShutdownTimeout uint   `env:"GRACEFUL_SHUTDOWN_TIMEOUT_SEC" env-default:"5"`
}

type LoggerConfig struct {
	IsProduction bool `env:"IS_PRODUCTION" env-default:"true"`
}

type SwaggerConfig struct {
	IsEnableSwagger bool   `env:"IS_ENABLE_SWAGGER" env-default:"true"`
	SwaggerURL      string `env:"SWAGGER_URL"`
}

type URLGeneratorConfig struct {
	BaseURL   string `env:"BASE_URL" env-required:"true"`
	JWTSecret string `env:"JWT_SECRET" env-required:"true"`
}

type Config struct {
	DBConfig
	HTTPConfig
	LoggerConfig
	SwaggerConfig
	URLGeneratorConfig
	StorageHandler  string `env:"STORAGE_HANDLER_URL" env-default:"http://0.0.0.0:3391/api/v1/get_models"`
	FeaturesHandler string `env:"FEATURES_HANDLER_URL" env-default:"http://0.0.0.0:3392/api/v1/face_model/save_features"`
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
