package internal

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/vertica/vertica-sql-go/logger"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
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

func (s *Service) CreateItem(ctx context.Context, p models.Item) (models.Item, error) {
	timeNow := time.Now()
	p.ID = uuid.New()
	p.CreatedAt = timeNow
	p.UpdatedAt = timeNow
	err := s.s.CreateItem(ctx, p)
	if err != nil {
		return models.Item{}, err
	}

	return s.s.GetItem(ctx, p.ID.String())
}

func (s *Service) GetItem(ctx context.Context, pID string) (models.Item, error) {
	poster, err := s.s.GetItem(ctx, pID)
	if err != nil {
		err = fiber.NewError(fiber.StatusNotFound, err.Error())

		return models.Item{}, err
	}

	return poster, nil
}

func (s *Service) DeleteItem(ctx context.Context, pID string) error {
	err := s.s.DeleteItem(ctx, pID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateItem(ctx context.Context, p models.Item) (models.Item, error) {
	timeNow := time.Now()
	p.UpdatedAt = timeNow
	err := s.s.UpdateItem(ctx, p)
	if err != nil {
		return models.Item{}, err
	}

	return s.GetItem(ctx, p.ID.String())
}

func (s *Service) SearchPosters(ctx context.Context, query string) ([]models.Item, error) {
	itemsSearch, err := s.s.SearchItems(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(itemsSearch) == 0 {
		return nil, nil
	}

	posters := make([]models.Item, 0, len(itemsSearch))

	for _, p := range itemsSearch {
		poster, err := s.s.GetItem(ctx, p.ID.String())
		if err != nil {
			return nil, err
		}

		posters = append(posters, poster)
	}

	return posters, nil
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
