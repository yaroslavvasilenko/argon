package parser

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/core/validator"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

func BodyParser(c *fiber.Ctx, out interface{}) error {
	err := c.BodyParser(out)
	if err != nil {
		return err
	}

	v := validator.Validate(out)
	if v.HasErrors() {
		return v.Error()
	}

	return nil
}

func ParamParser(c *fiber.Ctx, out interface{}) error {
	err := c.ParamsParser(out)
	if err != nil {
		return err
	}

	v := validator.Validate(out)
	if v.HasErrors() {
		return v.Error()
	}

	return nil
}

func QueryParser(c *fiber.Ctx, out interface{}) error {
	err := c.QueryParser(out)
	if err != nil {
		return err
	}

	v := validator.Validate(out)
	if v.HasErrors() {
		return v.Error()
	}

	return nil
}

func GetLang(ctx context.Context) string {
	return ctx.Value(models.KeyLanguage).(string)
}
