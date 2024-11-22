package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal"
)

const AppName = "argon"

func NewApiRouter(controllers *internal.Handler) *fiber.App {
	// Application (fiber)
	r := fiber.New(fiber.Config{
		//ErrorHandler:            handler.ErrorHandler,
		DisableStartupMessage:   false,
		AppName:                 AppName,
		EnableTrustedProxyCheck: false,
		//TrustedProxies:          cfg.App.TrustedProxies,
		//JSONEncoder:             JSONEncoder,
		//BodyLimit:               128 * 1024 * 1024,
	})

	r.Get("/ping", controllers.Ping)

	//  poster
	r.Post("/api/v1/poster", controllers.CreatePoster)
	r.Get("/api/v1/poster/:poster_id", controllers.GetPoster)
	r.Delete("/api/v1/poster/:poster_id", controllers.DeletePoster)
	r.Put("/api/v1/poster/:poster_id", controllers.UpdatePoster)

	return r
}
