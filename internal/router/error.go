package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/phuslu/log"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	// check fiber error
	if e, ok := err.(*fiber.Error); ok {
		switch e.Code {
		case fiber.StatusBadRequest:
			return c.Status(400).JSON(fiber.Map{
				"code":        "BadRequest",
				"description": e.Message,
			})
		case fiber.StatusNotFound:
			return c.Status(404).JSON(fiber.Map{
				"code":        "NotFound",
				"description": e.Message,
			})
		case fiber.StatusUnauthorized:
			return c.Status(401).JSON(fiber.Map{
				"code":        "Unauthorized",
				"description": "Недействительный токен аутентификации",
			})
		case fiber.StatusMethodNotAllowed:
			return c.Status(405).JSON(fiber.Map{
				"code":        "Method Not Allowed",
				"description": "Метод не поддерживается",
			})
		case fiber.StatusConflict:
			return c.Status(409).JSON(fiber.Map{
				"code":        "Conflict",
				"description": e.Message,
			})
		}
	}

	log.Error().Err(err).Msg("Internal Server Error")

	return c.Status(500).JSON(fiber.Map{
		"code":        "InternalServerError",
		"description": "Internal Server Error",
	})
}
