package internal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
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

func (h *Handler) CreateItem(c *fiber.Ctx) error {
	r := &struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}{}

	err := c.BodyParser(r)
	if err != nil {
		return err
	}

	poster, err := h.s.CreateItem(c.UserContext(), models.Item{
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(poster)
}

func (h *Handler) GetItem(c *fiber.Ctx) error {
	chatID := c.Params("item_id")

	item, err := h.s.GetItem(c.UserContext(), chatID)
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func (h *Handler) DeleteItem(c *fiber.Ctx) error {
	itemID := c.Params("item_id")

	err := h.s.DeleteItem(c.UserContext(), itemID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) UpdateItem(c *fiber.Ctx) error {
	itemID := uuid.UUID{}
	err := itemID.Scan(c.Params("item_id"))
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

	item, err := h.s.UpdateItem(c.UserContext(), models.Item{
		ID:    itemID,
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(item)
}

func (h *Handler) SearchItems(c *fiber.Ctx) error {
	q := new(struct {
		Query string `json:"query"`
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

func (h *Handler) GetCategories(c *fiber.Ctx) error {
	resp, err := h.s.GetCategories(c.UserContext())
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
