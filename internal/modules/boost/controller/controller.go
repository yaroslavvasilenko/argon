package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/modules/boost"
	"github.com/yaroslavvasilenko/argon/internal/modules/boost/service"
)

type Boost struct {
	s *service.Boost
}

func NewBoost(s *service.Boost) *Boost {
	return &Boost{s: s}
}

func (b *Boost) GetBoost(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("listing_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Проверяем, что UUID не пустой
	if id == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID листинга не может быть пустым",
		})
	}

	boost, err := b.s.GetBoost(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get boost",
		})
	}

	return c.JSON(boost)
}

func (b *Boost) UpdateBoost(c *fiber.Ctx) error {
	req := boost.UpdateBoostRequest{}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
		
	id, err := uuid.Parse(c.Params("listing_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	// Проверяем, что UUID не пустой
	if id == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID листинга не может быть пустым",
		})
	}

	req.ListingID = id

	boost, err := b.s.UpsertBoost(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update boost",
		})
	}

	return c.JSON(boost)
}
