package main

import (
	"context"
	"fmt"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/controller"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	"github.com/yaroslavvasilenko/argon/internal/router"
	"os"
)

// main initializes the application, loads environment variables from the .env file,
// creates the configuration app, creates a logger, and launches the application.
func main() {
	config.LoadConfig()

	cfg := config.GetConfig()

	lg := logger.NewLogger(cfg)

	lg.Info().Msg("starting app...")

	exit := func(msg string, err error) {
		lg.Err(err).Msg(msg)
		os.Exit(1)
	}

	ctx := context.Background()

	gorm, pool, err := db.NewSqlDB(ctx, cfg.DB.Url, lg.Logger, true)
	if err != nil {
		exit(fmt.Sprintf("connecting to db %s", cfg.DB.Url), err)
	}

	err = database.Migrate(cfg.DB.Url)
	if err != nil {
		exit(fmt.Sprintf("migrating db %s", cfg.DB.Url), err)
	}


	storagesDB := storage.NewStorage(gorm, pool)

	service := service.NewService(storagesDB)
	controller := controller.NewHandler(service)
	// init router
	r := router.NewApiRouter(controller)

	err = r.Listen(":" + cfg.Port)
	if err != nil {
		exit(fmt.Sprintf("starting server on port %s", cfg.Port), err)
	}

	return
}
