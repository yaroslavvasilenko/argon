package service

import (
	"context"

	"encoding/json"
	"errors"
	"fmt"
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

func (s *Listing) CreateListing(ctx context.Context, p listing.CreateListingRequest) (listing.CreateListingResponse, error) {
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
	}, p.Categories, p.Location, p.Characteristics)
	if err != nil {
		return listing.CreateListingResponse{}, err
	}

	fullListing, err := s.s.GetFullListing(ctx, ID.String())
	if err != nil {
		return listing.CreateListingResponse{}, err
	}

	boosts := []listing.BoostResp{}

	for _, boost := range fullListing.Boosts {
		boosts = append(boosts, listing.BoostResp{
			Type:              boost.Type,
			CommissionPercent: boost.Commission,
		})
	}

	resp := listing.CreateListingResponse{
		ID:          fullListing.Listing.ID,
		Title:       fullListing.Listing.Title,
		Description: fullListing.Listing.Description,
		Price:       fullListing.Listing.Price,
		Currency:    fullListing.Listing.Currency,
		Location:    fullListing.Location,
		Categories:  fullListing.Categories.ID,
		Characteristics: fullListing.Characteristics,
		Boosts:          boosts,
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

	boosts := []listing.BoostResp{}

	for _, boost := range fullListing.Boosts {
		boosts = append(boosts, listing.BoostResp{
			Type:              boost.Type,
			CommissionPercent: boost.Commission,
		})
	}

	resp := listing.FullListingResponse{
		ID:             fullListing.Listing.ID,
		Title:          fullListing.Listing.Title,
		Description:    fullListing.Listing.Description,
		Price:          fullListing.Listing.Price,
		Currency:       fullListing.Listing.Currency,
		Location:       fullListing.Location,
		Categories:     fullListing.Categories.ID,
		Characteristics: fullListing.Characteristics,
		Boosts:         boosts,
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
	}, p.Categories, p.Location, p.Characteristics)
	if err != nil {
		return listing.FullListingResponse{}, err
	}

	return s.GetListing(ctx, p.ID.String())
}

func (s *Listing) GetCategories(ctx context.Context) (listing.ResponseGetCategories, error) {
	// Получаем базовую структуру категорий
	var categories []listing.CategoryNode
	if err := json.Unmarshal([]byte(config.GetConfig().Categories.Json), &categories); err != nil {
		return listing.ResponseGetCategories{}, errors.New("ошибка при разборе базовых категорий: " + err.Error())
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
		return listing.ResponseGetCategories{}, errors.New("ошибка при разборе локализаций: " + err.Error())
	}

	// Рекурсивно применяем локализации
	applyTranslations(&categories, translations)

	return listing.ResponseGetCategories{Categories: categories}, nil
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


func (s *Listing) GetCharacteristicsForCategory(ctx context.Context, categoryIds []string) ([]models.CharacteristicItem, error) {
	// Получаем характеристики и переводы
	_, characteristicKeys, characteristicsTranslations, err := s.getCategoryCharacteristics(ctx, categoryIds)
	if err != nil {
		return nil, err
	}

	// Создаем массив характеристик в порядке их ключей
	result := make([]models.CharacteristicItem, 0, len(characteristicKeys))
	
	// Добавляем характеристики в порядке их следования в иерархии категорий
	for _, key := range characteristicKeys {
		if translation, ok := characteristicsTranslations[key]; ok {
			result = append(result, models.CharacteristicItem{
				Role:  key,
				Value: translation,
			})
		}
	}

	return result, nil
}

// GetFiltersForCategory возвращает фильтры для указанной категории
func (s *Listing) GetFiltersForCategory(ctx context.Context, categoryId string) (listing.GetFiltersForCategoryResponse, error) {
	// Используем общую функцию для получения характеристик категории
	// Передаем только одну категорию
	_, characteristicKeys, _, err := s.getCategoryCharacteristics(ctx, []string{categoryId})
	if err != nil {
		return listing.GetFiltersForCategoryResponse{}, err
	}

	// Получаем значения характеристик из БД
	charValues, err := s.s.GetCharacteristicValues(ctx, characteristicKeys)
	if err != nil {
		return listing.GetFiltersForCategoryResponse{}, fmt.Errorf("error getting characteristic values: %w", err)
	}

	// Просто возвращаем полученные значения характеристик
	return listing.GetFiltersForCategoryResponse{Filters: charValues}, nil
}

// getCategoryCharacteristics получает характеристики для указанных категорий
func (s *Listing) getCategoryCharacteristics(ctx context.Context, categoryIds []string) (map[string][]string, []string, map[string]string, error) {
	// Получаем язык из контекста
	lang := ctx.Value(models.KeyLanguage).(string)

	// Загружаем характеристики категорий из конфига
	var categoryCharacteristics map[string][]string
	if err := json.Unmarshal([]byte(config.GetConfig().Categories.Characteristics), &categoryCharacteristics); err != nil {
		return nil, nil, nil, errors.New("ошибка при разборе характеристик категорий: " + err.Error())
	}

	// Загружаем переводы характеристик в зависимости от языка
	var characteristicsTranslations map[string]string
	var translationsJson string

	// Выбираем нужный язык перевода
	switch lang {
	case string(models.LanguageRu):
		translationsJson = config.GetConfig().Categories.LangCharacteristics.Ru
	case string(models.LanguageEn):
		translationsJson = config.GetConfig().Categories.LangCharacteristics.En
	case string(models.LanguageEs):
		translationsJson = config.GetConfig().Categories.LangCharacteristics.Es
	default:
		translationsJson = config.GetConfig().Categories.LangCharacteristics.En
	}

	// Парсим переводы
	if err := json.Unmarshal([]byte(translationsJson), &characteristicsTranslations); err != nil {
		return nil, nil, nil, errors.New("ошибка при разборе переводов характеристик: " + err.Error())
	}

	// Создаем мапу для отслеживания уже добавленных характеристик
	characteristicsSet := make(map[string]bool)
	var characteristicKeys []string

	// Проходим по категориям в порядке их иерархии
	// Поскольку categoryIds уже содержит путь от корня до категории,
	// мы просто проходим по нему в том же порядке
	for _, categoryId := range categoryIds {
		// Получаем характеристики для текущей категории
		if chars, ok := categoryCharacteristics[categoryId]; ok {
			// Добавляем характеристики для текущей категории, сохраняя порядок
			for _, char := range chars {
				if !characteristicsSet[char] {
					characteristicsSet[char] = true
					characteristicKeys = append(characteristicKeys, char)
				}
			}
		}
	}

	return categoryCharacteristics, characteristicKeys, characteristicsTranslations, nil
}
