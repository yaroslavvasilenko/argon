package controller

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/modules/location"
	"github.com/yaroslavvasilenko/argon/internal/modules/location/service"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
)

type Location struct {
	s *service.Location
}

func NewLocation(s *service.Location) *Location {
	return &Location{s: s}
}

func (h *Location) GetLocation(c *fiber.Ctx) error {
	req := location.GetLocationRequest{}
	if err := parser.BodyParser(c, &req); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	locResp, err := h.s.GetLocation(c.UserContext(), req)
	if err != nil {
		return err
	}

	return c.JSON(locResp)
}
