package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/modules"
)

const AppName = "argon"

func NewApiRouter(controllers *modules.Controllers) *fiber.App {
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

	r.Get("/ping", controllers.Listing.Ping)

	//  poster
	r.Post("/api/v1/listing", controllers.Listing.CreateListing)
	r.Get("/api/v1/listing/:listing_id", controllers.Listing.GetListing)
	r.Delete("/api/v1/listing/:listing_id", controllers.Listing.DeleteListing)
	r.Put("/api/v1/listing/:listing_id", controllers.Listing.UpdateListing)

	//  search
	r.Post("/api/v1/search", controllers.Listing.SearchListings)

	//  categories
	r.Get("/api/v1/categories", controllers.Listing.GetCategories)

	//  currency
	r.Get("/api/v1/currency", controllers.Currency.GetCurrency)
	
	return r
}
