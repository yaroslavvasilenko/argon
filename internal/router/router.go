package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/yaroslavvasilenko/argon/internal/auth"
	authmw "github.com/yaroslavvasilenko/argon/internal/auth/middleware"
	"github.com/yaroslavvasilenko/argon/internal/middleware"
	"github.com/yaroslavvasilenko/argon/internal/modules"
)

const AppName = "argon"

// NewApiRouter now takes authSvc and wires the JWT check on /api/v1/*
func NewApiRouter(
	controllers *modules.Controllers,
	authSvc auth.Service,
) *fiber.App {
	r := fiber.New(fiber.Config{
		ErrorHandler:            ErrorHandler,
		DisableStartupMessage:   false,
		AppName:                 AppName,
		EnableTrustedProxyCheck: false,
		//TrustedProxies:          cfg.App.TrustedProxies,
		//JSONEncoder:             JSONEncoder,
		//BodyLimit:               128 * 1024 * 1024,
	})

	// CORS & Language (unprotected)
	r.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	r.Use(middleware.Language())
	// introspect users everywhere
	r.Use(authmw.AuthenticationMiddleware(authSvc))

	// health‑check, open to all
	r.Get("/ping", controllers.Listing.Ping)

	// JWT auth middleware
	protected := r.Group("", authmw.RequireAuth())

	//  ── Listing ─────────────────────────────────────────
	r.Get("/listing/:listing_id", controllers.Listing.GetListing)
	protected.Post("/listing", controllers.Listing.CreateListing)
	protected.Put("/listing/:listing_id", controllers.Listing.UpdateListing)
	protected.Delete("/listing/:listing_id", controllers.Listing.DeleteListing)

	//  ── Search ──────────────────────────────────────────
	r.Post("/search", controllers.Listing.SearchListings)
	r.Get("/search/params", controllers.Listing.SearchListingsParams)

	//  ── Categories ──────────────────────────────────────
	r.Get("/categories", controllers.Listing.GetCategories)
	r.Post("/categories/characteristics", controllers.Listing.GetCharacteristicsForCategory)
	r.Get("/categories/filters", controllers.Listing.GetFiltersForCategory)

	//  ── Currency & Location ─────────────────────────────
	r.Get("/currency", controllers.Currency.GetCurrency)
	r.Post("/location", controllers.Location.GetLocation)

	//  ── Boost ───────────────────────────────────────────
	protected.Post("/boost/:listing_id", controllers.Boost.UpdateBoost)
	protected.Get("/boost/:listing_id", controllers.Boost.GetBoost)

	//  ── Images ──────────────────────────────────────────
	protected.Post("/images/upload", controllers.Image.UploadImage)
	r.Get("/images/get/:image_id", controllers.Image.GetImage)

	return r
}
