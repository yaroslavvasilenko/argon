package parser

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

func BodyParser(c *fiber.Ctx, out interface{}) error {
	// Получаем значение заголовка Accept-Language с дефолтным значением es
	acceptLang := c.Get(models.HeaderLanguage, models.LanguageDefault)

	// Устанавливаем язык в контекст запроса
	c.Locals(models.KeyLanguage, acceptLang)

	return c.BodyParser(out)
}

func GetLang(ctx context.Context) string {
	// Получаем значение заголовка Accept-Language с дефолтным значением es
	acceptLang := ctx.Value(models.KeyLanguage).(string)

	return acceptLang
}
