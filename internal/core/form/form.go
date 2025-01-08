package form

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func parseQuery(c *fiber.Ctx, out interface{}) error {
	if err := c.QueryParser(out); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := validate(out).Error(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return nil
}

func parseParams(c *fiber.Ctx, out interface{}) error {
	if err := c.ParamsParser(out); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := validate(out).Error(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return nil
}

func parseBody(c *fiber.Ctx, out interface{}) error {
	if err := c.BodyParser(out); err != nil {
		var e *json.UnmarshalTypeError
		if errors.As(err, &e) {
			return NewValidationError(map[string][]string{e.Field: {
				fmt.Sprintf("field must be %s", e.Type.String()),
			}})
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return validate(out).Error()
}
