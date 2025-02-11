package listing

import (
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// SearchListingsRequest represents a request to search listings.
type SearchListingsRequest struct {
	// Query is the search query.
	Query string `json:"query" query:"query" validate:"required"`
	// Limit is the maximum number of results to return.
	Limit int `json:"limit,omitempty" query:"limit" validate:"omitempty,min=-100,max=100"` // default: 20
	// Cursor is the cursor for pagination.
	Cursor string `json:"cursor,omitempty" query:"cursor"`
	// SortOrder is the order in which to sort the results.
	SortOrder string `json:"sort_order,omitempty" query:"sort_order" validate:"omitempty,oneof=price_asc price_desc relevance_asc relevance_desc popularity_asc popularity_desc"`
	// SearchID is the ID of the search.
	SearchID string `json:"qid,omitempty" query:"search_id,omitempty"`
	// Filters are the filters to apply to the search.
	Filters Filters `json:"filters,omitempty"`
	// Category is the category to search in.
	Category string `json:"category,omitempty"`
	// currency: SupportedCurrency;
	// locale: string;
	// location?: Location;
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
