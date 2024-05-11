package logger

import (
	"go.uber.org/zap"
	"log"
)

type LoggerConfig struct {
	IsProduction bool
}

func NewLogger(cfg LoggerConfig) *zap.Logger {
	var l *zap.Logger
	var err error
	if cfg.IsProduction {
		l, err = zap.NewProduction()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		l, err = zap.NewDevelopment()
		if err != nil {
			log.Fatal(err)
		}
	}

	return l
}
