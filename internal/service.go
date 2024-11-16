package internal

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/entity"
	"time"
)

type Service struct {
	s *Storage
}

func NewService(s *Storage) *Service {
	return &Service{s: s}
}

func (s *Service) Ping() string {
	return "pong"
}

func (s *Service) CreatePoster(p entity.Poster) error {
	timeNow := time.Now()
	p.ID = uuid.New()
	p.CreatedAt = timeNow
	p.UpdatedAt = timeNow
	err := s.s.CreatePoster(p)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetPoster(pID string) (entity.Poster, error) {
	poster, err := s.s.GetPoster(pID)
	if err != nil {
		return entity.Poster{}, err
	}

	return poster, nil
}

func (s *Service) DeletePoster(pID string) error {
	err := s.s.DeletePoster(pID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdatePoster(p entity.Poster) error {
	timeNow := time.Now()
	p.UpdatedAt = timeNow
	err := s.s.UpdatePoster(p)
	if err != nil {
		return err
	}

	return nil
}
