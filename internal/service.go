package internal

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
	"gorm.io/gorm"
)

type Service struct {
	s *Storage
	logger.Logger
}

func NewService(s *Storage) *Service {
	return &Service{s: s}
}

func (s *Service) Ping() string {
	return "pong"
}

func (s *Service) CreateListing(ctx context.Context, p models.Listing) (models.Listing, error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()

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

func (s *Service) SearchListings(ctx context.Context, query string) ([]models.Listing, error) {
	listingsSearch, err := s.s.SearchListings(ctx, query)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return nil, err
	}

	listings := make([]models.Listing, 0, len(listingsSearch))

	for _, p := range listingsSearch {
		listing, err := s.s.GetListing(ctx, p.ID.String())
		if err != nil {
			return nil, err
		}

		listings = append(listings, listing)
	}

	return listings, nil
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
