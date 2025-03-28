package listing

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// SearchListingsResponse represents a response to a search listings request.

type ResponseGetCategories struct {
	Categories []CategoryNode `json:"categories"`
}

type CategoryNode struct {
	Category      Category       `json:"category"`
	Subcategories []CategoryNode `json:"subcategories,omitempty"`
}

type Category struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Image *string `json:"image,omitempty"`
}

type CreateListingRequest struct {
	Title           string                `json:"title"`
	Description     string                `json:"description,omitempty"`
	Price           float64               `json:"price,omitempty"`
	Currency        models.Currency       `json:"currency,omitempty"`
	Location        models.Location       `json:"location,omitempty"`
	Categories      []string              `json:"categories,omitempty"`
	Characteristics models.Characteristic `json:"characteristics,omitempty"`
	Images          []string              `json:"images"`
}

func GetCreateListingRequest(c *fiber.Ctx) (CreateListingRequest, error) {
	req := CreateListingRequest{}
	err := c.BodyParser(&req)
	if err != nil {
		return CreateListingRequest{}, err
	}

	validCategoryIds := config.GetConfig().Categories.CategoryIds
	for _, categoryId := range req.Categories {
		if !validCategoryIds[categoryId] {
			return CreateListingRequest{}, errors.New("invalid category ID: " + categoryId)
		}
	}

	return req, nil
}

type CreateListingResponse struct {
	ID              uuid.UUID              `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description,omitempty"`
	Price           float64                `json:"price,omitempty"`
	Currency        models.Currency        `json:"currency,omitempty"`
	Location        models.Location        `json:"location,omitempty"`
	Categories      []Category             `json:"categories"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts          []BoostResp            `json:"boosts,omitempty"`
	Images          []string               `json:"images"`
}

type BoostResp struct {
	Type              models.BoostType `json:"type"`
	CommissionPercent float64          `json:"commission_percent"`
}

type UpdateListingRequest struct {
	ID              uuid.UUID              `json:"id" validate:"required"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description,omitempty"`
	Price           float64                `json:"price,omitempty" validate:"gte=0"`
	Currency        models.Currency        `json:"currency,omitempty" validate:"required,oneof=USD EUR RUB"`
	Location        models.Location        `json:"location,omitempty"`
	Categories      []string               `json:"categories,omitempty" validate:"required"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts          []BoostResp            `json:"boosts,omitempty"`
	Images          []string               `json:"images"`
}

type FullListingResponse struct {
	ID                  uuid.UUID             `json:"id"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	OriginalDescription string                `json:"original_description"`
	Price               float64               `json:"price"`
	Currency            models.Currency       `json:"currency"`
	OriginalPrice       float64               `json:"original_price"`
	OriginalCurrency    models.Currency       `json:"original_currency"`
	Location            models.Location       `json:"location"`
	Seller              models.Seller         `json:"seller"`
	Categories          []Category            `json:"categories"`
	Characteristics     models.Characteristic `json:"characteristics"`
	Images              []string              `json:"images"`
	CreatedAt           int64                `json:"created_at"`
	UpdatedAt           int64                `json:"updated_at"`
	Boosts              []BoostResp           `json:"boosts,omitempty"`
	IsEditable          bool                  `json:"is_editable"`
	IsBuyable           bool                  `json:"is_buyable"`
	IsNSFW              bool                  `json:"is_nsfw"`
}

type GetFiltersForCategoryResponse struct {
	Filters models.Filters `json:"filter_params"`
}

// getCategoriesWithLocalizedNames получает локализованные названия категорий
func GetCategoriesWithLocalizedNames(ctx context.Context, categoryIDs []string) ([]Category, error) {
	// Получаем язык из контекста
	lang := models.Localization(ctx.Value(models.KeyLanguage).(string))
	if lang == "" {
		lang = models.LanguageDefault
	}

	// Загружаем локализации
	var translations map[string]string
	var langData string

	switch lang {
	case models.LanguageRu:
		langData = config.GetConfig().Categories.Lang.Ru
	case models.LanguageEn:
		langData = config.GetConfig().Categories.Lang.En
	case models.LanguageEs:
		langData = config.GetConfig().Categories.Lang.Es
	default:
		langData = config.GetConfig().Categories.Lang.Es
	}

	// Распаковываем локализации
	if err := json.Unmarshal([]byte(langData), &translations); err != nil {
		return nil, errors.New("ошибка при разборе локализаций: " + err.Error())
	}

	// Создаем массив категорий с локализованными названиями
	categories := make([]Category, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		name, ok := translations[categoryID]
		if !ok {
			name = categoryID // Если перевод не найден, используем ID как имя
		}
		categories = append(categories, Category{
			ID:   categoryID,
			Name: name,
		})
	}

	return categories, nil
}
