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
	r.Post("/api/v1/listing", controllers.CreateListing)
	r.Get("/api/v1/listing/:listing_id", controllers.GetListing)
	r.Delete("/api/v1/listing/:listing_id", controllers.DeleteListing)
	r.Put("/api/v1/listing/:listing_id", controllers.UpdateListing)

	//  search
	r.Get("/api/v1/search", controllers.SearchListings)

	//  categories
	r.Get("/api/v1/categories", controllers.GetCategories)

	return r
}
