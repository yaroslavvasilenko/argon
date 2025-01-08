package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"gorm.io/gorm"

	"encoding/base64"
)

type Service struct {
	s *storage.Storage
	logger.Logger
}

func NewService(s *storage.Storage) *Service {
	return &Service{s: s}
}

func (s *Service) Ping() string {
	return "pong"
}

func (s *Service) CreateListing(ctx context.Context, p models.Listing) (models.Listing, error) {
	p.ID = uuid.New()
	timeNow := time.Now()
	p.CreatedAt = timeNow
	p.UpdatedAt = timeNow

	err := s.s.CreateListing(ctx, p)
	if err != nil {
		return models.Listing{}, err
	}
	return s.s.GetListing(ctx, p.ID.String())
}

func (s *Service) GetListing(ctx context.Context, pID string) (models.Listing, error) {
	listing, err := s.s.GetListing(ctx, pID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Listing{}, err
		}
		return models.Listing{}, err
	}

	return listing, nil
}

func (s *Service) DeleteListing(ctx context.Context, pID string) error {
	err := s.s.DeleteListing(ctx, pID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	return nil
}

func (s *Service) UpdateListing(ctx context.Context, p models.Listing) (models.Listing, error) {
	p.UpdatedAt = time.Now()

	err := s.s.UpdateListing(ctx, p)
	if err != nil {
		return models.Listing{}, err
	}

	return s.GetListing(ctx, p.ID.String())
}

func (s *Service) SearchListings(ctx context.Context, req listing.SearchListingsRequest) (listing.SearchListingsResponse, error) {
	var cursor listing.SearchCursor
	if req.Cursor != "" {
		cursorBytes, err := base64.StdEncoding.DecodeString(req.Cursor)
		if err != nil {
			return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
		}
		if err := json.Unmarshal(cursorBytes, &cursor); err != nil {
			return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusBadRequest, "invalid cursor format")
		}
	}

	resp := listing.SearchListingsResponse{}
	var searchTitle, searchDescription bool

	if cursor.Block == "" || cursor.Block == listing.TitleBlock {
		listings, err := s.s.SearchListingsByTitle(ctx, req.Query, req.Limit, cursor.LastIndex)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}
		resp.Results = listings
		searchTitle = true
	}

	if cursor.Block == listing.DescriptionBlock || len(resp.Results) < req.Limit {
		if cursor.Block != listing.DescriptionBlock {
			cursor.LastIndex = nil
		}

		listings, err := s.s.SearchListingsByDescription(ctx, req.Query, req.Limit - len(resp.Results), cursor.LastIndex)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return listing.SearchListingsResponse{}, fiber.NewError(fiber.StatusNotFound, err.Error())
			}
			return listing.SearchListingsResponse{}, err
		}
		resp.Results = append(resp.Results, listings...)
		searchDescription = true
	}

	//  cursor
	//  cursorAfter
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

		cursorBytes, err := json.Marshal(newCursor)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}
		resp.CursorAfter = base64.StdEncoding.EncodeToString(cursorBytes)
	}

	// cursorBefore
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

		cursorBytes, err := json.Marshal(newCursor)
		if err != nil {
			return listing.SearchListingsResponse{}, err
		}
		resp.CursorBefore = base64.StdEncoding.EncodeToString(cursorBytes)
	}

	return resp, nil
}

func (s *Service) GetCategories(ctx context.Context) (map[string]interface{}, error) {
	var catMap map[string]interface{}

	// Преобразуем CategoriesJson из конфига в map[string]interface{}
	err := json.Unmarshal([]byte(config.GetConfig().CategoriesJson), &catMap)
	if err != nil {
		return catMap, err
	}

	return catMap, nil
}
