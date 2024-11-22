package internal

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/entity"
	"gorm.io/gorm"
	"time"
)

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

const posterTable = "posters"

func NewStorage(db *gorm.DB, pool *pgxpool.Pool) *Storage {
	return &Storage{gorm: db, pool: pool}
}

func (s *Storage) CreatePoster(p entity.Poster) error {
	err := s.gorm.Create(&p).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetPoster(pID string) (entity.Poster, error) {
	poster := entity.Poster{}

	err := s.gorm.Table(posterTable).
		First(&poster).
		Where("id = ?, deleted_at IS NULL", pID).
		Error
	if err != nil {
		return entity.Poster{}, err
	}

	return poster, nil
}

func (s *Storage) DeletePoster(pID string) error {
	err := s.gorm.Table(posterTable).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdatePoster(p entity.Poster) error {
	err := s.gorm.Updates(&p).
		Where("deleted_at IS NULL").
		Error
	if err != nil {
		return err
	}
	return nil
}
