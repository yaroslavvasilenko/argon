package internal

import "github.com/gofiber/fiber/v2"

type Handler struct {
	s *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{s: s}
}

func (h *Handler) Ping(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"ping": h.s.Ping(),
	})

}
