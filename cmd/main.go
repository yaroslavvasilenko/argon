package main

import (
	"context"
	"fmt"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/router"
)

// main initializes the application, loads environment variables from the .env file,
// creates the configuration app, creates a logger, and launches the application.
func main() {
	config.LoadConfig()

	cfg := config.GetConfig()

	ctx := context.Background()

	gorm, pool, err := db.NewSqlDB(ctx, cfg.DB.Url)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = database.Migrate(cfg.DB.Url)
	if err != nil {
		fmt.Println(err)
		return
	}

	storagesDB := internal.NewStorage(gorm, pool)

	service := internal.NewService(storagesDB)
	controller := internal.NewHandler(service)
	// init router
	r := router.NewApiRouter(controller)

	err = r.Listen(":" + cfg.Port)
	if err != nil {
		fmt.Println(err)
	}

	return
}
