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

	poster, err := h.s.CreatePoster(c.UserContext(), entity.Poster{
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(poster)
}

func (h *Handler) GetPoster(c *fiber.Ctx) error {
	chatID := c.Params("poster_id")

	poster, err := h.s.GetPoster(c.UserContext(), chatID)
	if err != nil {
		return err
	}

	return c.JSON(poster)
}

func (h *Handler) DeletePoster(c *fiber.Ctx) error {
	chatID := c.Params("poster_id")

	err := h.s.DeletePoster(c.UserContext(), chatID)
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

	poster, err := h.s.UpdatePoster(c.UserContext(), entity.Poster{
		ID:    posterID,
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(poster)
}

func (h *Handler) SearchPosters(c *fiber.Ctx) error {
	q := new(struct {
		Query string `json:"query" validate:"required,min=1,max=64"`
	})

	if err := c.QueryParser(q); err != nil {
		return err
	}

	posters, err := h.s.SearchPosters(c.UserContext(), q.Query)
	if err != nil {
		return err
	}

	return c.JSON(posters)
}
