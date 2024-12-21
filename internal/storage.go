package internal

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/opensearch"
	"gorm.io/gorm"
	"time"
)

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
	os   *opensearch.OpenSearch
}

const itemTable = "items"

func NewStorage(db *gorm.DB, pool *pgxpool.Pool, os *opensearch.OpenSearch) *Storage {
	return &Storage{gorm: db, pool: pool, os: os}
}

func (s *Storage) CreateItem(ctx context.Context, p models.Item) error {
	err := s.gorm.Create(&p).WithContext(ctx).Error
	if err != nil {
		return err
	}

	go s.os.IndexItems(ctx, []models.Item{p})

	return nil
}

func (s *Storage) GetItem(ctx context.Context, pID string) (models.Item, error) {
	poster := models.Item{}

	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		First(&poster).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Item{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return models.Item{}, err
	}

	return poster, nil
}

func (s *Storage) DeleteItem(ctx context.Context, pID string) error {
	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	go s.os.DeleteItem(ctx, pID)

	return nil
}

func (s *Storage) UpdateItem(ctx context.Context, p models.Item) error {
	err := s.gorm.Updates(&p).WithContext(ctx).
		Where("deleted_at IS NULL").
		Error
	if err != nil {
		return err
	}

	go s.os.IndexItems(ctx, []models.Item{p})

	return nil
}

func (s *Storage) SearchItems(ctx context.Context, query string) ([]models.ItemSearch, error) {
	return s.os.SearchItems(ctx, query)
}
