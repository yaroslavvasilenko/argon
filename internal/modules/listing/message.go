package listing

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// SearchListingsRequest represents a request to search listings.
type SearchListingsRequest struct {
	Query string `json:"query" query:"query" validate:"required"`
	Limit int `json:"limit,omitempty" query:"limit" validate:"omitempty,min=-100,max=100"` // default: 20
	Cursor string `json:"cursor,omitempty" query:"cursor"`
	SortOrder string `json:"sort_order,omitempty" query:"sort_order" validate:"omitempty,oneof=price_asc price_desc relevance_asc relevance_desc popularity_asc popularity_desc"`
	SearchID string `json:"qid,omitempty" query:"search_id,omitempty"`
	Filters models.Characteristics `json:"filters,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
	Currency models.Currency `json:"currency,omitempty"`
	LocationID string `json:"location_id,omitempty"`
}

// SearchListingsResponse represents a response to a search listings request.
type SearchListingsResponse struct {
	// Results are the search results.
	Results []models.Listing `json:"results"`
	// CursorAfter is the cursor for the next page of results.
	CursorAfter string `json:"cursor_after,omitempty"`
	// CursorBefore is the cursor for the previous page of results.
	CursorBefore string `json:"cursor_before,omitempty"`
	// SearchID is the ID of the search.
	SearchID string `json:"search_id,omitempty" query:"search_id,omitempty"`
}

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
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price,omitempty"`
	Currency    models.Currency `json:"currency,omitempty"`
	Location    models.Location `json:"location,omitempty"`
	Categories  []string        `json:"categories,omitempty"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
}

type CreateListingResponse struct {
	Title       string          `json:"title" validate:"required"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price,omitempty" validate:"gte=0"`
	Currency    models.Currency `json:"currency,omitempty" validate:"required,oneof=USD EUR RUB"`
	Location    models.Location `json:"location,omitempty"`
	Categories  []string        `json:"categories,omitempty" validate:"required"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts      []BoostResp     `json:"boosts,omitempty"`
}

type BoostResp struct {
	Type              models.BoostType `json:"type"`
	CommissionPercent float64          `json:"commission_percent"`
}

type UpdateListingRequest struct {
	ID          uuid.UUID       `json:"id" validate:"required"`
	Title       string          `json:"title" validate:"required"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price,omitempty" validate:"gte=0"`
	Currency    models.Currency `json:"currency,omitempty" validate:"required,oneof=USD EUR RUB"`
	Location    models.Location `json:"location,omitempty"`
	Categories  []string        `json:"categories,omitempty" validate:"required"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts      []BoostResp     `json:"boosts,omitempty"`
}

type FullListingResponse struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price,omitempty"`
	Currency    models.Currency `json:"currency,omitempty"`
	Location    models.Location `json:"location,omitempty"`
	Categories  []string        `json:"categories,omitempty"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts      []BoostResp     `json:"boosts,omitempty"`
}
