package main

import (
	"context"
	"fmt"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/opensearch"
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
		lg.Err(err).Msg("exiting")
		os.Exit(1)
	}

	ctx := context.Background()

	gorm, pool, err := db.NewSqlDB(ctx, cfg.DB.Url)
	if err != nil {
		exit(fmt.Sprintf("connecting to db %s", cfg.DB.Url), err)
	}

	err = database.Migrate(cfg.DB.Url)
	if err != nil {
		exit(fmt.Sprintf("migrating db %s", cfg.DB.Url), err)
	}

	openSearch, err := opensearch.NewOpenSearch(cfg.OpenSearch.Addr, cfg.OpenSearch.Login, cfg.OpenSearch.Password, cfg.OpenSearch.PosterIndex)
	if err != nil {
		exit(fmt.Sprintf("connecting to opensearch %+v", cfg.OpenSearch), err)
	}

	storagesDB := internal.NewStorage(gorm, pool, openSearch)

	service := internal.NewService(storagesDB)
	controller := internal.NewHandler(service)
	// init router
	r := router.NewApiRouter(controller)

	err = r.Listen(":" + cfg.Port)
	if err != nil {
		exit(fmt.Sprintf("starting server on port %s", cfg.Port), err)
	}

	return
}
