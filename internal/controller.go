package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/entity"
)

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

func (h *Handler) CreatePoster(c *fiber.Ctx) error {
	r := &struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}{}

	err := c.BodyParser(r)
	if err != nil {
		return err
	}

	err = h.s.CreatePoster(entity.Poster{
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetPoster(c *fiber.Ctx) error {
	chatID := c.Params("chat_id")

	poster, err := h.s.GetPoster(chatID)
	if err != nil {
		return err
	}

	return c.JSON(poster)
}

func (h *Handler) DeletePoster(c *fiber.Ctx) error {
	chatID := c.Params("poster_id")

	err := h.s.DeletePoster(chatID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) UpdatePoster(c *fiber.Ctx) error {
	posterID := uuid.UUID{}
	err := posterID.Scan(c.Params("poster_id"))
	if err != nil {
		return err
	}

	r := &struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}{}

	err = c.BodyParser(r)
	if err != nil {
		return err
	}

	err = h.s.UpdatePoster(entity.Poster{
		ID:    posterID,
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return nil
}
