package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// Language middleware добавляет информацию о языке в контекст запроса
func Language() fiber.Handler {
	return func(c *fiber.Ctx) error {	
		// Получаем язык из заголовка
		lang := c.Get(models.HeaderLanguage, models.LanguageDefault)

		// Проверяем, что язык поддерживается
		_, ok := models.LocalMap[models.Localization(lang)]
		if !ok {
			return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
				"error": "don't support language: " + string(lang),
			})
		}
		
		// Устанавливаем язык в контекст
		c.Locals(models.KeyLanguage, lang)
		
		// Продолжаем обработку запроса
		return c.Next()
	}
}
