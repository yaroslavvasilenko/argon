package internal

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/yaroslavvasilenko/argon/internal/entity"
	"time"
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

func (s *Service) CreatePoster(ctx context.Context, p entity.Poster) (entity.Poster, error) {
	timeNow := time.Now()
	p.ID = uuid.New()
	p.CreatedAt = timeNow
	p.UpdatedAt = timeNow
	err := s.s.CreatePoster(ctx, p)
	if err != nil {
		return entity.Poster{}, err
	}

	return s.s.GetPoster(ctx, p.ID.String())
}

func (s *Service) GetPoster(ctx context.Context, pID string) (entity.Poster, error) {
	poster, err := s.s.GetPoster(ctx, pID)
	if err != nil {
		err = fiber.NewError(fiber.StatusNotFound, err.Error())

		return entity.Poster{}, err
	}

	return poster, nil
}

func (s *Service) DeletePoster(ctx context.Context, pID string) error {
	err := s.s.DeletePoster(ctx, pID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdatePoster(ctx context.Context, p entity.Poster) (entity.Poster, error) {
	timeNow := time.Now()
	p.UpdatedAt = timeNow
	err := s.s.UpdatePoster(ctx, p)
	if err != nil {
		return entity.Poster{}, err
	}

	return s.GetPoster(ctx, p.ID.String())
}

func (s *Service) SearchPosters(ctx context.Context, query string) ([]entity.Poster, error) {
	postersSearch, err := s.s.SearchPosters(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(postersSearch) == 0 {
		return nil, nil
	}

	posters := make([]entity.Poster, 0, len(postersSearch))

	for _, p := range postersSearch {
		fmt.Println(p.ID.String())
		poster, err := s.s.GetPoster(ctx, p.ID.String())
		if err != nil {
			return nil, err
		}

		posters = append(posters, poster)
	}

	return posters, nil
}
