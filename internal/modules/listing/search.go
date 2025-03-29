package listing

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

type SearchListingsRequest struct {
	Query      string          `json:"query" query:"query"`
	Currency   models.Currency `json:"currency,omitempty"`
	Cursor     string          `json:"cursor,omitempty"`
	SearchID   string          `json:"qid,omitempty"`
	Limit      int             `json:"limit,omitempty"`
	CategoryID string          `json:"category_id,omitempty"`
	Location   models.Location `json:"location,omitempty"`
	Filters    models.Filters  `json:"filters,omitempty"`
	SortOrder  string          `json:"sort_order,omitempty"`
}

func GetSearchListingsRequest(c *fiber.Ctx) (SearchListingsRequest, error) {
	req := struct {
		Query      *string         `json:"query"`
		Currency   models.Currency `json:"currency,omitempty"`
		Cursor     string          `json:"cursor,omitempty"`
		SearchID   string          `json:"qid,omitempty"`
		Limit      int             `json:"limit,omitempty"`
		CategoryID string          `json:"category_id,omitempty"`
		Location   models.Location `json:"location,omitempty"`
		Filters    models.Filters  `json:"filters,omitempty"`
		SortOrder  string          `json:"sort_order,omitempty"`
	}{}

	if err := c.BodyParser(&req); err != nil {
		return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest, "Error parsing request body: "+err.Error())
	}

	// Validate required fields
	if req.Query == nil {
		return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest, "query parameter is required")
	}

	// Validate currency
	if err := req.Currency.Validate(); err != nil {
		return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Validate limit
	if req.Limit == 0 {
		// Set default limit if not provided
		req.Limit = 20
	} else if req.Limit < -100 || req.Limit > 100 {
		return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest,
			fmt.Sprintf("limit must be between -100 and 100, got: %d", req.Limit))
	}

	// Validate sort order if provided
	if req.SortOrder != "" {
		sortOrderValid := false
		for _, so := range models.SortOrders {
			if req.SortOrder == so {
				sortOrderValid = true
				break
			}
		}
		if !sortOrderValid {
			return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest,
				fmt.Sprintf("invalid sort_order: %s. Supported values: %v", req.SortOrder, models.SortOrders))
		}
	}

	// Validate location if provided
	if req.Location.ID != "" {
		if err := req.Location.Validate(); err != nil {
			return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	// Validate filters if provided
	if len(req.Filters) > 0 {
		if err := req.Filters.Validate(); err != nil {
			return SearchListingsRequest{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	}

	return SearchListingsRequest{
		Query:      *req.Query,
		Currency:   req.Currency,
		Cursor:     req.Cursor,
		SearchID:   req.SearchID,
		Limit:      req.Limit,
		CategoryID: req.CategoryID,
		Location:   req.Location,
		Filters:    req.Filters,
		SortOrder:  req.SortOrder,
	}, nil
}

type SearchListingsResponse struct {
	Results      []ListingResponse `json:"items"`
	CursorAfter  *string           `json:"cursor_after"`
	CursorBefore *string           `json:"cursor_before"`
	SearchID     string            `json:"qid"`
}

type ListingResponse struct {
	ItemID           uuid.UUID       `json:"item_id"`
	Title            string          `json:"title"`
	Price            float64         `json:"price"`
	Currency         models.Currency `json:"currency"`
	OriginalPrice    float64         `json:"original_price"`
	OriginalCurrency models.Currency `json:"original_currency"`
	Description      string          `json:"description"`
	Location         models.Location `json:"location"`
	Category         Category        `json:"category"`
	Images           []string        `json:"images"`
	IsHighlighted    bool            `json:"is_highlighted"`
	IsBuyable        bool            `json:"is_buyable"`
}

// CreateSearchListingsResponse создает ответ на запрос поиска объявлений
func CreateSearchListingsResponse(
	ctx context.Context,
	listings []models.ListingResult,
	cursorAfter *string,
	cursorBefore *string,
	searchID string,
) (SearchListingsResponse, error) {
	results := make([]ListingResponse, 0, len(listings))

	for _, listingResult := range listings {
		listing := listingResult.Listing

		// Подготавливаем данные для ответа
		var categoryInfo Category
		var isHighlighted bool
		var isBuyable bool
		var location models.Location

		// Обрабатываем категории
		if len(listingResult.Categories) > 0 {
			// Получаем локализованное название категории
			categoryID := listingResult.Categories[0].ID[0]
			categoryName := ""
			categoryNames, err := GetCategoriesWithLocalizedNames(ctx, []string{categoryID})
			if err != nil {
				return SearchListingsResponse{}, err
			}

			if len(categoryNames) != 0 {
				categoryName = categoryNames[0].Name
			}
			categoryInfo = Category{
				ID:   categoryID,
				Name: categoryName,
			}
		}

		// Обрабатываем локацию
		if listingResult.Location.ID != "" {
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
			Images:           []string{},
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
	}, nil
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
	Filters    models.Filters
	SortOrder  string
	LocationID string
}

type GetSearchParamsResponse struct {
	Category *Category        `json:"category,omitempty"`
	Location *models.Location `json:"location,omitempty"`
	Filters  *models.Filters  `json:"filters,omitempty"`
	SortOrder *string         `json:"sort_order,omitempty"`
}
