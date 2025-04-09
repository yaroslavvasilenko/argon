package middleware

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// Language middleware добавляет информацию о языке в контекст запроса
func Language() fiber.Handler {
	return func(c *fiber.Ctx) error {
		lang := models.Localization(c.Get(models.HeaderLanguage, string(models.LanguageDefault)))

		_, ok := models.LocalMap[lang]
		if !ok {
			slog.WarnContext(c.UserContext(), "don't support language: "+string(lang))
			lang = models.LanguageDefault
		}

		// Устанавливаем язык в Go-контекст
		ctx := context.WithValue(c.UserContext(), models.KeyLanguage, lang)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
