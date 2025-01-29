package service

import (
	"context"

	"gorm.io/gorm"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

func (s *Service) SearchListings(ctx context.Context, req listing.SearchListingsRequest) (listing.SearchListingsResponse, error) {
	var cursor listing.SearchCursor
	var err error
	if req.Cursor != "" && req.SearchID != "" {
		cursor, err = s.cache.GetCursor(req.Cursor)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}

		search, err := s.cache.GetSearchInfo(req.SearchID)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}

		req.SortOrder = search.SortOrder
		req.Filters = search.Filters
	}

	resp := listing.SearchListingsResponse{}
	var searchTitle, searchDescription bool
	uniqueListings := make(map[uuid.UUID]struct{}) // Для отслеживания уникальных листингов

	if cursor.Block == "" || cursor.Block == listing.TitleBlock {
		listings, err := s.s.SearchListingsByTitle(ctx, req.Query, req.Limit, cursor.LastIndex, req.SortOrder)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}

		// Добавляем только уникальные листинги
		for _, l := range listings {
			if _, exists := uniqueListings[l.ID]; !exists {
				uniqueListings[l.ID] = struct{}{}
				resp.Results = append(resp.Results, l)
			}
		}
		searchTitle = true
	}

	if cursor.Block == listing.DescriptionBlock || len(resp.Results) < req.Limit {
		if cursor.Block != listing.DescriptionBlock {
			cursor.LastIndex = nil
		}

		remainingLimit := req.Limit - len(resp.Results)
		listings, err := s.s.SearchListingsByDescription(ctx, req.Query, remainingLimit, cursor.LastIndex, req.SortOrder)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}

		// Добавляем только уникальные листинги из поиска по описанию
		for _, l := range listings {
			if _, exists := uniqueListings[l.ID]; !exists {
				uniqueListings[l.ID] = struct{}{}
				resp.Results = append(resp.Results, l)
			}
		}
		searchDescription = true
	}

	if len(resp.Results) > 0 && len(resp.Results) == req.Limit {
		lastListing := resp.Results[len(resp.Results)-1]

		newCursor := listing.SearchCursor{
			LastIndex: &lastListing.ID,
		}

		if searchDescription {
			newCursor.Block = listing.DescriptionBlock
		} else if searchTitle {
			newCursor.Block = listing.TitleBlock
		}

		resp.CursorAfter = s.cache.StoreCursor(newCursor)
	}

	if len(resp.Results) > 0 && req.Cursor != "" {
		firstListing := resp.Results[0]

		newCursor := listing.SearchCursor{
			LastIndex: &firstListing.ID,
		}

		if searchTitle {
			newCursor.Block = listing.TitleBlock
		} else if searchDescription {
			newCursor.Block = listing.DescriptionBlock
		}

		resp.CursorBefore = s.cache.StoreCursor(newCursor)
	}

	searchId := listing.SearchId{
		Category:  req.Category,
		Filters:   req.Filters,
		SortOrder: req.SortOrder,
	}

	resp.SearchID = s.cache.StoreSearchInfo(searchId)

	return resp, nil
}
