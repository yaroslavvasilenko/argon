package controller

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
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
	r := &listing.CreateListingRequest{}

	err := c.BodyParser(r)
	if err != nil {
		return err
	}

	// Валидируем категории
	validCategoryIds := config.GetConfig().Categories.CategoryIds
	for _, categoryId := range r.Categories {
		if !validCategoryIds[categoryId] {
			return errors.New("invalid category ID: " + categoryId)
		}
	}

	listing, err := h.s.CreateListing(c.UserContext(), r)
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
		return fiber.NewError(fiber.StatusBadRequest, "Неверный формат ID листинга")
	}
	
	// Проверяем, что UUID не пустой
	if listingID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID листинга не может быть пустым")
	}

	r := listing.UpdateListingRequest{}

	err = c.BodyParser(&r)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Ошибка при разборе тела запроса: "+err.Error())
	}
	
	// Устанавливаем ID из параметра URL в запрос
	r.ID = listingID

	listing, err := h.s.UpdateListing(c.UserContext(), r)
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
	lang := c.Get(models.HeaderLanguage, models.LanguageDefault)

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, lang)

	resp, err := h.s.GetCategories(ctx)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
