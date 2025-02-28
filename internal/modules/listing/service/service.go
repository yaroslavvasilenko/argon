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

func (s *Listing) CreateListing(ctx context.Context, p *listing.CreateListingRequest) (listing.CreateListingResponse, error) {
	ID := uuid.New()
	timeNow := time.Now()

	err := s.s.CreateListing(ctx, models.Listing{
		ID:          ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		Currency:    p.Currency,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}, p.Categories, p.Location)
	if err != nil {
		return listing.CreateListingResponse{}, err
	}

	fullListing, err := s.s.GetFullListing(ctx, ID.String())
	if err != nil {
		return listing.CreateListingResponse{}, err
	}

	resp := listing.CreateListingResponse{
		Title:       fullListing.Listing.Title,
		Description: fullListing.Listing.Description,
		Price:       fullListing.Listing.Price,
		Currency:    fullListing.Listing.Currency,
		Location:    fullListing.Location,
		Categories:  fullListing.Categories.ID,
	}

	return resp, nil
}

func (s *Listing) GetListing(ctx context.Context, pID string) (listing.FullListingResponse, error) {
	fullListing, err := s.s.GetFullListing(ctx, pID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return listing.FullListingResponse{}, err
		}
		return listing.FullListingResponse{}, err
	}

	resp := listing.FullListingResponse{
		ID:          fullListing.Listing.ID,
		Title:       fullListing.Listing.Title,
		Description: fullListing.Listing.Description,
		Price:       fullListing.Listing.Price,
		Currency:    fullListing.Listing.Currency,
		Location:    fullListing.Location,
		Categories:  fullListing.Categories.ID,
	}

	return resp, nil
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

func (s *Listing) UpdateListing(ctx context.Context, p listing.UpdateListingRequest) (listing.FullListingResponse, error) {
	// Устанавливаем ListingID в Location
	p.Location.ListingID = p.ID

	err := s.s.UpdateFullListing(ctx, models.Listing{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		Currency:    p.Currency,
		UpdatedAt:   time.Now(),
	}, p.Categories, p.Location)
	if err != nil {
		return listing.FullListingResponse{}, err
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
