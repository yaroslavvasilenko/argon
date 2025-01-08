package listing

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// type SearchFilters struct {
// 	// Добавьте необходимые фильтры
// }

type SearchListingsRequest struct {
	Query   string `json:"query" validate:"required"`
	Limit   int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"` // default: 20
	Cursor  string `json:"cursor,omitempty"`
	// Filters SearchFilters
}

type SearchListingsResponse struct {
	Results      []models.Listing `json:"results"`
	CursorAfter  string           `json:"cursor_after,omitempty"`
	CursorBefore string           `json:"cursor_before,omitempty"`
}

type SearchBlock string

const (
	TitleBlock       SearchBlock = "title"
	DescriptionBlock SearchBlock = "description"
)

type SearchCursor struct {
	Block     SearchBlock `json:"b"`
	LastIndex *uuid.UUID  `json:"i"`
}
