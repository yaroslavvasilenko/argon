package controller

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	r, err := listing.GetCreateListingRequest(c)
	if err != nil {
		return err
	}

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, c.Get(models.HeaderLanguage, models.LanguageDefault))

	listing, err := h.s.CreateListing(ctx, r)
	if err != nil {
		return err
	}

	return c.JSON(listing)
}

func (h *Listing) GetListing(c *fiber.Ctx) error {
	listingID := c.Params("listing_id")

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, c.Get(models.HeaderLanguage, models.LanguageDefault))

	listing, err := h.s.GetListing(ctx, listingID)
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
	err := c.BodyParser(&req)
	if err != nil {
		return err
	}

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, c.Get(models.HeaderLanguage, models.LanguageDefault))

	listings, err := h.s.SearchListings(ctx, req)
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

func (h *Listing) SearchListingsParams(c *fiber.Ctx) error {
	qID := c.Query("qid")
	if qID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "qid parameter is required")
	}

	listings, err := h.s.GetSearchParams(c.UserContext(), qID)
	if err != nil {
		return err
	}

	return c.JSON(listings)
}

func (h *Listing) GetCharacteristicsForCategory(c *fiber.Ctx) error {
	// Новая структура запроса с полем category_ids
	req := struct {
		CategoryIds []string `json:"category_ids"`
	}{}

	err := c.BodyParser(&req)
	if err != nil {
		return err
	}

	// Получаем язык из заголовка Accept-Language, по умолчанию используем "en"
	lang := c.Get(models.HeaderLanguage, models.LanguageDefault)

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, lang)

	// Вызываем сервис с новым форматом данных
	characteristics, err := h.s.GetCharacteristicsForCategory(ctx, req.CategoryIds)
	if err != nil {
		return err
	}

	return c.JSON(characteristics)
}

// GetFiltersForCategory возвращает фильтры для указанной категории
func (h *Listing) GetFiltersForCategory(c *fiber.Ctx) error {
	// Получаем category_id из параметров запроса
	categoryId := c.Query("category_id")
	if categoryId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category_id is required",
		})
	}

	// Получаем язык из заголовка Accept-Language, по умолчанию используем "en"
	lang := c.Get(models.HeaderLanguage, models.LanguageDefault)

	// Создаем контекст с информацией о языке
	ctx := context.WithValue(c.UserContext(), models.KeyLanguage, lang)

	// Вызываем сервис для получения фильтров
	filters, err := h.s.GetFiltersForCategory(ctx, categoryId)
	if err != nil {
		return err
	}

	return c.JSON(filters)
}
