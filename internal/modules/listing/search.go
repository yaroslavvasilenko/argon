package listing

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

type SearchListingsRequest struct {
	Query      string          `json:"query" query:"query" validate:"required"`
	Currency   models.Currency `json:"currency,omitempty"`
	Cursor     string          `json:"cursor,omitempty" query:"cursor"`
	SearchID   string          `json:"qid,omitempty" query:"search_id,omitempty"`
	Limit      int             `json:"limit,omitempty" query:"limit" validate:"omitempty,min=-100,max=100"`
	CategoryID string          `json:"category_id,omitempty"`
	Location   models.Location `json:"location,omitempty"`
	Filters    Filters         `json:"filters,omitempty"`
	SortOrder  string          `json:"sort_order,omitempty" query:"sort_order" validate:"omitempty,oneof=price_asc price_desc relevance_asc relevance_desc popularity_asc popularity_desc"`
}

type SearchListingsResponse struct {
	Results      []ListingResponse `json:"items"`
	CursorAfter  string            `json:"cursor_after,omitempty"`
	CursorBefore string            `json:"cursor_before,omitempty"`
	SearchID     string            `json:"search_id,omitempty" query:"search_id,omitempty"`
}

type ListingResponse struct {
	ItemID           uuid.UUID       `json:"item_id"`
	Title            string          `json:"title"`
	Price            float64         `json:"price"`
	Currency         models.Currency `json:"currency"`
	OriginalPrice    float64         `json:"original_price,omitempty"`
	OriginalCurrency models.Currency `json:"original_currency,omitempty"`
	Description      string          `json:"description,omitempty"`
	Location         models.Location `json:"location,omitempty"`
	Category         CategoryInfo    `json:"category,omitempty"`
	Images           []string        `json:"images,omitempty"`
	IsHighlighted    bool            `json:"is_highlighted,omitempty"`
	IsBuyable        bool            `json:"is_buyable,omitempty"`
}

type CategoryInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
}

// CreateSearchListingsResponse создает ответ на запрос поиска объявлений
func CreateSearchListingsResponse(
	listings []models.ListingResult,
	cursorAfter string,
	cursorBefore string,
	searchID string,
) SearchListingsResponse {
	results := make([]ListingResponse, 0, len(listings))

	for _, listingResult := range listings {
		listing := listingResult.Listing

		// Подготавливаем данные для ответа
		var categoryInfo CategoryInfo
		var isHighlighted bool
		var isBuyable bool
		var location models.Location

		// Обрабатываем категории
		if len(listingResult.Categories) > 0 {
			// TODO: сделать получение имени и изображения категории из справочника категорий
			categoryInfo = CategoryInfo{
				ID:   listingResult.Categories[0].ID[0],
				Name: "TODO: получить имя категории",
			}
		}

		// Обрабатываем локацию
		if listingResult.Location.ID != uuid.Nil {
			location = listingResult.Location
		}

		// Обрабатываем бусты
		for _, boost := range listingResult.Boosts {
			if boost.Type == models.BoostTypeHighlight {
				isHighlighted = true
			}
			// TODO: сделать логику для определения IsBuyable на основе бустов
			if boost.Type == models.BoostTypeUpfront {
				isBuyable = true
			}
		}

		// TODO: обработка изображений
		// В будущем здесь будет код для получения изображений

		// Создаем модель ответа в конце, после сбора всех данных
		response := ListingResponse{
			ItemID:           listing.ID,
			Title:            listing.Title,
			Price:            listing.Price,
			Currency:         listing.Currency,
			OriginalPrice:    listing.Price,
			OriginalCurrency: listing.Currency,
			Description:      listing.Description,
			Location:         location,
			Category:         categoryInfo,
			IsHighlighted:    isHighlighted,
			IsBuyable:        isBuyable,
			// Можно добавить характеристики и изображения, если они нужны в ответе
		}

		results = append(results, response)
	}

	return SearchListingsResponse{
		Results:      results,
		CursorAfter:  cursorAfter,
		CursorBefore: cursorBefore,
		SearchID:     searchID,
	}
}

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
	Filters    Filters
	SortOrder  string
	LocationID string
}

