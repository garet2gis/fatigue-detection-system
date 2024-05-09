package main

import (
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/cmd/commands/model_trainer"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/cmd/commands/serve"
	_ "github.com/garet2gis/fatigue-detection-system/model_handler_service/docs"
	_ "github.com/garet2gis/fatigue-detection-system/model_handler_service/migrations"

	"github.com/urfave/cli/v2"
	"log"
	"os"
)

//	@title		Model storage service
//	@version	1.0

//	@BasePath	/api/v1/

func main() {
	app := &cli.App{
		Name: "data-service",
		Commands: []*cli.Command{
			{
				Name:   "serve",
				Action: serve.Action,
			},
			{
				Name:   "model-trainer",
				Action: model_trainer.Action,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
