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

func (h *Handler) CreateListing(c *fiber.Ctx) error {
	r := &struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}{}

	err := c.BodyParser(r)
	if err != nil {
		return err
	}

	listing, err := h.s.CreateListing(c.UserContext(), models.Listing{
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Handler) GetListing(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	listing, err := h.s.GetListing(c.UserContext(), listingID)
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Handler) DeleteListing(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	err := h.s.DeleteListing(c.UserContext(), listingID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) UpdateListing(c *fiber.Ctx) error {
	listingID := uuid.UUID{}
	err := listingID.Scan(c.Params("listing_id"))
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

	listing, err := h.s.UpdateListing(c.UserContext(), models.Listing{
		ID:    listingID,
		Title: r.Title,
		Text:  r.Text,
	})
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Handler) SearchListings(c *fiber.Ctx) error {
	q := new(struct {
		Query string `json:"query"`
	})

	if err := c.QueryParser(q); err != nil {
		return err
	}

	listings, err := h.s.SearchListings(c.UserContext(), q.Query)
	if err != nil {
		return err
	}

	return c.JSON(listings)
}

func (h *Handler) GetCategories(c *fiber.Ctx) error {
	resp, err := h.s.GetCategories(c.UserContext())
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
