package middleware

import (
	"context"
	"github.com/gofiber/fiber/v2"

	"github.com/yaroslavvasilenko/argon/internal/models"
)

// Language middleware puts language into context for later use
func Language() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lang := models.Localization(c.Get(models.HeaderLanguage, string(models.LanguageDefault)))

		_, ok := models.LocalMap[lang]
		if !ok {
			lang = models.LanguageDefault
		}

		ctx := context.WithValue(c.UserContext(), models.KeyLanguage, lang)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
