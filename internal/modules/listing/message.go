package listing

import (
	"errors"
	"time"

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
	Categories      []string               `json:"categories,omitempty"`
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
	ID                  uuid.UUID              `json:"id"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	OriginalDescription string                 `json:"original_description"`
	Price               float64                `json:"price"`
	Currency            models.Currency        `json:"currency"`
	OriginalPrice       float64                `json:"original_price"`
	OriginalCurrency    models.Currency        `json:"original_currency"`
	Location            models.Location        `json:"location,omitempty"`
	Seller              models.Seller          `json:"seller,omitempty"`
	Categories          []string               `json:"categories"`
	Characteristics     map[string]interface{} `json:"characteristics"`
	Images              []string               `json:"images"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	Boosts              []BoostResp            `json:"boosts,omitempty"`
	IsEditable          bool                 `json:"is_editable,omitempty"`
	IsBuyout            bool                 `json:"is_buyout,omitempty"`
	IsNSFW              bool                 `json:"is_nsfw,omitempty"`
}
type GetFiltersForCategoryResponse struct {
	Filters models.Filters `json:"filter_params"`
}
