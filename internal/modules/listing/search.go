package listing

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

type SearchBlock string

const (
	TitleBlock       SearchBlock = "title"
	DescriptionBlock SearchBlock = "description"
)

type SearchCursor struct {
	Block     SearchBlock
	LastIndex *uuid.UUID
}

type SearchID struct {
	CategoryID string
	Filters    models.Characteristics
	SortOrder  string
	LocationID string
}

type PriceFilterParams struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type ColorFilterParams struct {
	Color string `json:"color,omitempty"`
}

type FilterType string

const (
	PriceFilter FilterType = "Price"
	ColorFilter FilterType = "Color"
)

type Filter struct {
	Type  FilterType        `json:"type"`
	Color ColorFilterParams `json:"color,omitempty"`
	Price PriceFilterParams `json:"price,omitempty"`
}

type Filters []Filter
