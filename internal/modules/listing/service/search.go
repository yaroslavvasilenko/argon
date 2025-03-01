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

		req.SortOrder = search.SortOrder
		req.Filters = search.Filters
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

	if cursor.Block == "" || cursor.Block == listing.TitleBlock {
		var listings []models.Listing
		listingAnchor, listings, err = s.s.SearchListingsByTitle(ctx, req.Query, req.Limit, cursor.LastIndex, req.SortOrder)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}

		resp.Results = append(resp.Results, listings...)
		searchTitle = true
	}

	if cursor.Block == listing.DescriptionBlock || len(resp.Results) < req.Limit {
		if cursor.Block != listing.DescriptionBlock {
			cursor.LastIndex = nil
		}

		remainingLimit := req.Limit - len(resp.Results)
		//  TODO: сделать исключение по запросу перенести запрос Title
		listings, err := s.s.SearchListingsByDescription(ctx, req.Query, remainingLimit, cursor.LastIndex, req.SortOrder)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}

		resp.Results = append(resp.Results, listings...)
		searchDescription = true
	}

	if len(resp.Results) > 0 && len(resp.Results) == int(math.Abs(float64(req.Limit))) {
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


func (s *Listing) GetSearchParams(ctx context.Context, qID string) (listing.SearchId, error) {
	search, err := s.cache.GetSearchInfo(qID)
	if err != nil {
		return listing.SearchId{}, err
	}

	return search, nil
}