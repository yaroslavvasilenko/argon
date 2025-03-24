package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// Добавляем CORS middleware для разрешения запросов с любых адресов
	r.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length, Content-Type",
	}))

	r.Get("/ping", controllers.Listing.Ping)

	//  poster
	r.Post("/api/v1/listing", controllers.Listing.CreateListing)
	r.Get("/api/v1/listing/:listing_id", controllers.Listing.GetListing)
	r.Delete("/api/v1/listing/:listing_id", controllers.Listing.DeleteListing)
	r.Put("/api/v1/listing/:listing_id", controllers.Listing.UpdateListing)

	//  search
	r.Post("/api/v1/search", controllers.Listing.SearchListings)
	r.Get("/api/v1/search/params", controllers.Listing.SearchListingsParams)

	//  categories
	r.Get("/api/v1/categories", controllers.Listing.GetCategories)
	r.Post("/api/v1/categories/characteristics", controllers.Listing.GetCharacteristicsForCategory)
	r.Get("/api/v1/categories/filters", controllers.Listing.GetFiltersForCategory)

	//  currency
	r.Get("/api/v1/currency", controllers.Currency.GetCurrency)

	//  location
	r.Post("/api/v1/location", controllers.Location.GetLocation)

	//  boost
	r.Post("/api/v1/boost/:listing_id", controllers.Boost.UpdateBoost)
	r.Get("/api/v1/boost/:listing_id", controllers.Boost.GetBoost)

	return r
}
