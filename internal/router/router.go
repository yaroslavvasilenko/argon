package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal"
)

const AppName = "argon"

func NewApiRouter(controllers *internal.Handler) *fiber.App {
	// Application (fiber)
	r := fiber.New(fiber.Config{
		ErrorHandler:            ErrorHandler,
		DisableStartupMessage:   false,
		AppName:                 AppName,
		EnableTrustedProxyCheck: false,
		//TrustedProxies:          cfg.App.TrustedProxies,
		//JSONEncoder:             JSONEncoder,
		//BodyLimit:               128 * 1024 * 1024,
	})

	r.Get("/ping", controllers.Ping)

	//  poster
	r.Post("/api/v1/item", controllers.CreateItem)
	r.Get("/api/v1/item/:item_id", controllers.GetItem)
	r.Delete("/api/v1/item/:item_id", controllers.DeleteItem)
	r.Put("/api/v1/item/:item_id", controllers.UpdateItem)

	//  search
	r.Get("/api/v1/search", controllers.SearchItems)

	//  categories
	r.Get("/api/v1/categories", controllers.GetCategories)

	return r
}
