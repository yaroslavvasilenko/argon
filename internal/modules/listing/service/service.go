package service

import (
	"context"

	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/config"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"gorm.io/gorm"
)

type Service struct {
	s      *storage.Storage
	logger *logger.LogPhuslu
	cache  *storage.Cache
}

func NewService(s *storage.Storage, pool *pgxpool.Pool, logger *logger.LogPhuslu) *Service {
	srv := &Service{
		s:      s,
		cache:  storage.NewCache(pool),
		logger: logger,
	}

	return srv
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

func (s *Service) GetCategories(ctx context.Context) (map[string]interface{}, error) {
	var catMap map[string]interface{}

	// Преобразуем CategoriesJson из конфига в map[string]interface{}
	err := json.Unmarshal([]byte(config.GetConfig().CategoriesJson), &catMap)
	if err != nil {
		return catMap, err
	}

	return catMap, nil
}
