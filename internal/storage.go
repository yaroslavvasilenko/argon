package internal

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
	"time"
)

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

const itemTable = "listings"

func NewStorage(db *gorm.DB, pool *pgxpool.Pool) *Storage {
	return &Storage{gorm: db, pool: pool}
}

func (s *Storage) CreateListing(ctx context.Context, p models.Listing) error {
	err := s.gorm.Create(&p).WithContext(ctx).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetListing(ctx context.Context, pID string) (models.Listing, error) {
	poster := models.Listing{}

	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		First(&poster).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Listing{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return models.Listing{}, err
	}

	return poster, nil
}

func (s *Storage) DeleteListing(ctx context.Context, pID string) error {
	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateListing(ctx context.Context, p models.Listing) error {
	err := s.gorm.Updates(&p).WithContext(ctx).
		Where("deleted_at IS NULL").
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SearchListings(ctx context.Context, query string) ([]models.ListingSearch, error) {
	return []models.ListingSearch{}, nil
}
