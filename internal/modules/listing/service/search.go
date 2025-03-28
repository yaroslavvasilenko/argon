package service

import (
	"context"
	"math"

	"gorm.io/gorm"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

func (s *Listing) SearchListings(ctx context.Context, req listing.SearchListingsRequest) (listing.SearchListingsResponse, error) {
	var cursor listing.SearchCursor
	var err error

	if req.SearchID != "" {
		search, err := s.cache.GetSearchInfo(req.SearchID)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}

		if search != nil {
			req.SortOrder = search.SortOrder
			req.Filters = search.Filters
		}
	}

	if req.Cursor != "" {
		cursor, err = s.cache.GetCursor(req.Cursor)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}
	}

	resp := listing.SearchListingsResponse{}
	var searchTitle, searchDescription bool
	var listingAnchor *models.Listing

	// Используем абсолютное значение для емкости слайса, чтобы избежать ошибки при отрицательном значении req.Limit
	listingsRes := make([]models.ListingResult, 0, int(math.Abs(float64(req.Limit))))
	if cursor.Block == "" || cursor.Block == listing.TitleBlock {
		var listings []models.ListingResult
		listingAnchor, listings, err = s.s.SearchListingsByTitle(ctx, req.Query, req.Limit, cursor.LastIndex, req.SortOrder, req.CategoryID, req.Filters, req.Location)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}

		listingsRes = append(listingsRes, listings...)
		searchTitle = true
	}

	if cursor.Block == listing.DescriptionBlock || len(listingsRes) < req.Limit {
		if cursor.Block != listing.DescriptionBlock {
			cursor.LastIndex = nil
		}

		// remainingLimit := req.Limit - len(resp.Results)
		//  TODO: сделать исключение по запросу перенести запрос Title
		// listings, err := s.s.SearchListingsByDescription(ctx, req.Query, remainingLimit, cursor.LastIndex, req.SortOrder)
		// if err != nil {
		// 	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 		return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		// 	}
		// 	return listing.SearchListingsResponse{}, err
		// }

		// resp.Results = append(resp.Results, listings...)
		// searchDescription = true
	}

	if len(listingsRes) > 0 && len(listingsRes) == int(math.Abs(float64(req.Limit))) {
		lastListing := listingsRes[len(listingsRes)-1]

		newCursor := listing.SearchCursor{
			LastIndex: &lastListing.Listing.ID,
		}

		if searchDescription {
			newCursor.Block = listing.DescriptionBlock
		} else if searchTitle {
			newCursor.Block = listing.TitleBlock
		}

		cursor := s.cache.StoreCursor(newCursor)

		resp.CursorAfter = &cursor
	}

	if listingAnchor != nil {
		firstListing := listingAnchor

		newCursor := listing.SearchCursor{
			LastIndex: &firstListing.ID,
		}

		if searchTitle {
			newCursor.Block = listing.TitleBlock
		} else if searchDescription {
			newCursor.Block = listing.DescriptionBlock
		}

		cursor := s.cache.StoreCursor(newCursor)

		resp.CursorBefore = &cursor
	}

	searchId := listing.SearchID{
		CategoryID:  req.CategoryID,
		Filters:   req.Filters,
		SortOrder: req.SortOrder,
	}

	resp.SearchID = s.cache.StoreSearchInfo(searchId)

	

	return listing.CreateSearchListingsResponse(ctx, listingsRes, 
		resp.CursorAfter, resp.CursorBefore, resp.SearchID)
}


func (s *Listing) GetSearchParams(ctx context.Context, qID string) (listing.SearchID, error) {
	search, err := s.cache.GetSearchInfo(qID)
	if err != nil {
		return listing.SearchID{}, err
	}

	if search == nil {
		return listing.SearchID{}, nil
	}

	return *search, nil
}
