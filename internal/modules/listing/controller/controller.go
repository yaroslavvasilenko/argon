package controller

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	"github.com/go-playground/validator/v10"
)

type Listing struct {
	s *service.Listing
}

func NewListing(s *service.Listing) *Listing {
	return &Listing{s: s}
}

func (h *Listing) Ping(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"ping": h.s.Ping(),
	})

}

func (h *Listing) CreateListing(c *fiber.Ctx) error {
	r := &struct {
		Title    string          `json:"title" validate:"required"`
		Description string          `json:"description" validate:"required"`
		Price    float64         `json:"price" validate:"required,gte=0"`
		Currency models.Currency `json:"currency" validate:"required,oneof=USD EUR RUB"`
	}{}

	err := c.BodyParser(r)
	if err != nil {
		return err
	}

	listing, err := h.s.CreateListing(c.UserContext(), models.Listing{
		Title:       r.Title,
		Description: r.Description,
		Price:       r.Price,
		Currency:    r.Currency,
	})
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Listing) GetListing(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	listing, err := h.s.GetListing(c.UserContext(), listingID)
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Listing) DeleteListing(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	err := h.s.DeleteListing(c.UserContext(), listingID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Listing) UpdateListing(c *fiber.Ctx) error {
	listingID := uuid.UUID{}
	err := listingID.Scan(c.Params("listing_id"))
	if err != nil {
		return err
	}

	r := &struct {
		Title    string          `json:"title" validate:"required"`
		Text     string          `json:"text" validate:"required"`
		Price    float64         `json:"price" validate:"required,gte=0"`
		Currency models.Currency `json:"currency" validate:"required,oneof=USD EUR RUB"`
	}{}

	err = c.BodyParser(r)
	if err != nil {
		return err
	}

	listing, err := h.s.UpdateListing(c.UserContext(), models.Listing{
		ID:          listingID,
		Title:       r.Title,
		Description: r.Text,
		Price:       r.Price,
		Currency:    r.Currency,
	})
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Listing) SearchListings(c *fiber.Ctx) error {
	req := listing.SearchListingsRequest{}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	if req.Query == "" {	
		return fiber.NewError(fiber.StatusBadRequest, "query parameter is required")
	}

	listings, err := h.s.SearchListings(c.UserContext(), req)
	if err != nil {
		return err
	}

	return c.JSON(listings)
}

func (h *Listing) GetCategories(c *fiber.Ctx) error {
	// Получаем язык из заголовка Accept-Language, по умолчанию используем "en"
	lang := c.Get("Accept-Language", "en")
	
	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), "lang", lang)
	
	
	resp, err := h.s.GetCategories(ctx)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
