package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules"
	"github.com/yaroslavvasilenko/argon/internal/router"

)

// main initializes the application, loads environment variables from the .env file,
// creates the configuration app, creates a logger, and launches the application.
func main() {
	config.LoadConfig()

	cfg := config.GetConfig()

	lg := logger.NewLogger(cfg)

	lg.Infof("starting app...")

	exit := func(msg string, err error) {
		lg.Errorf("%s", msg)
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

	storages := modules.NewStorages(cfg, gorm, pool)
	services := modules.NewServices(storages, pool, lg)
	controller := modules.NewControllers(services)
	// init router
	r := router.NewApiRouter(controller)

	err = r.Listen(":" + cfg.Port)
	if err != nil {
		exit(fmt.Sprintf("starting server on port %s", cfg.Port), err)
	}

	return
}
