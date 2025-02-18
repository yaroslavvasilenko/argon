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
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"gorm.io/gorm"
)

type Listing struct {
	s      *storage.Listing
	logger *logger.Glog
	cache  *storage.Cache
}

func NewListing(s *storage.Listing, pool *pgxpool.Pool, logger *logger.Glog) *Listing {
	srv := &Listing{
		s:      s,
		cache:  storage.NewCache(pool),
		logger: logger,
	}

	return srv
}

func (s *Listing) Ping() string {
	return "pong"
}

func (s *Listing) CreateListing(ctx context.Context, p models.Listing) (models.Listing, error) {
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

func (s *Listing) GetListing(ctx context.Context, pID string) (models.Listing, error) {
	listing, err := s.s.GetListing(ctx, pID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Listing{}, err
		}
		return models.Listing{}, err
	}

	return listing, nil
}

func (s *Listing) DeleteListing(ctx context.Context, pID string) error {
	err := s.s.DeleteListing(ctx, pID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return err
	}

	return nil
}

func (s *Listing) UpdateListing(ctx context.Context, p models.Listing) (models.Listing, error) {
	p.UpdatedAt = time.Now()

	err := s.s.UpdateListing(ctx, p)
	if err != nil {
		return models.Listing{}, err
	}

	return s.GetListing(ctx, p.ID.String())
}

func (s *Listing) GetCategories(ctx context.Context) ([]listing.CategoryNode, error) {
	// Получаем базовую структуру категорий
	var categories []listing.CategoryNode
	if err := json.Unmarshal([]byte(config.GetConfig().Categories.Json), &categories); err != nil {
		return nil, errors.New("ошибка при разборе базовых категорий: " + err.Error())
	}

	// Получаем язык из контекста
	lang := models.Localization(ctx.Value(models.KeyLanguage).(string))
	if lang == "" {
		lang = models.LanguageDefault
	}

	// Загружаем локализации
	var translations map[string]string
	var langData string

	switch lang {
	case models.LanguageRu:
		langData = config.GetConfig().Categories.Lang.Ru
	case models.LanguageEn:
		langData = config.GetConfig().Categories.Lang.En
	case models.LanguageEs:
		langData = config.GetConfig().Categories.Lang.Es
	default:
		s.logger.Warnf("неизвестный язык %s, используем дефолтный", lang)
		langData = config.GetConfig().Categories.Lang.Es
	}

	// Распаковываем локализации
	if err := json.Unmarshal([]byte(langData), &translations); err != nil {
		return nil, errors.New("ошибка при разборе локализаций: " + err.Error())
	}

	// Рекурсивно применяем локализации
	applyTranslations(&categories, translations)

	return categories, nil
}

// applyTranslations рекурсивно применяет переводы к категориям
func applyTranslations(nodes *[]listing.CategoryNode, translations map[string]string) {
	for i := range *nodes {
		node := &(*nodes)[i]
		if name, ok := translations[node.Category.ID]; ok {
			node.Category.Name = name
		}
		if len(node.Subcategories) > 0 {
			applyTranslations(&node.Subcategories, translations)
		}
	}
}
