package main

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/cmd/commands/model_trainer"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/cmd/commands/model_updater"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/cmd/commands/serve"
	"github.com/urfave/cli/v2"
	"log"
	"os"

	_ "github.com/garet2gis/fatigue-detection-system/user_data_service/docs"
	_ "github.com/garet2gis/fatigue-detection-system/user_data_service/migrations"
)

//	@title		User data service API
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
				Name:   "model-updater",
				Action: model_updater.Action,
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
