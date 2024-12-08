package internal

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/entity"
	"github.com/yaroslavvasilenko/argon/internal/opensearch"
	"gorm.io/gorm"
	"time"
)

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
	os   *opensearch.OpenSearch
}

const posterTable = "posters"

func NewStorage(db *gorm.DB, pool *pgxpool.Pool, os *opensearch.OpenSearch) *Storage {
	return &Storage{gorm: db, pool: pool, os: os}
}

func (s *Storage) CreatePoster(ctx context.Context, p entity.Poster) error {
	err := s.gorm.Create(&p).WithContext(ctx).Error
	if err != nil {
		return err
	}

	go s.os.IndexPosters(ctx, []entity.Poster{p})

	return nil
}

func (s *Storage) GetPoster(ctx context.Context, pID string) (entity.Poster, error) {
	poster := entity.Poster{}

	err := s.gorm.Table(posterTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		First(&poster).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.Poster{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return entity.Poster{}, err
	}

	return poster, nil
}

func (s *Storage) DeletePoster(ctx context.Context, pID string) error {
	err := s.gorm.Table(posterTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	go s.os.DeletePoster(ctx, pID)

	return nil
}

func (s *Storage) UpdatePoster(ctx context.Context, p entity.Poster) error {
	err := s.gorm.Updates(&p).WithContext(ctx).
		Where("deleted_at IS NULL").
		Error
	if err != nil {
		return err
	}

	go s.os.IndexPosters(ctx, []entity.Poster{p})

	return nil
}

func (s *Storage) SearchPosters(ctx context.Context, query string) ([]entity.PosterSearch, error) {
	return s.os.SearchPosters(ctx, query)
}
