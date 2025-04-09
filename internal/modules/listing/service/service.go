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
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"github.com/yaroslavvasilenko/argon/internal/modules/location/service"
	"gorm.io/gorm"
)

type Listing struct {
	s      *storage.Listing
	logger *logger.Glog
	cache  *storage.Cache
	location *service.Location
}

func NewListing(s *storage.Listing, pool *pgxpool.Pool, logger *logger.Glog, locationService *service.Location) *Listing {
	srv := &Listing{
		s:      s,
		cache:  storage.NewCache(pool),
		logger: logger,
		location: locationService,
	}

	return srv
}

func (s *Listing) Ping() string {
	return "pong"
}

func (s *Listing) CreateListing(ctx context.Context, p listing.CreateListingRequest) (listing.FullListingResponse, error) {
	ID := uuid.New()
	timeNow := time.Now()

	// Создаем объявление с переданными ID категорий
	err := s.s.CreateListing(ctx, models.Listing{
		ID:          ID,
		Title:       p.Title,
		Description: p.Description,
		Price:       p.Price,
		Currency:    p.Currency,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}, p.Categories, *p.Location, p.Characteristics)
	if err != nil {
		return listing.FullListingResponse{}, err
	}

	resp, err := s.GetListing(ctx, ID.String())
	if err != nil {
		return listing.FullListingResponse{}, err
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

	// Получаем локализованные названия категорий
	categories, err := listing.GetCategoriesWithLocalizedNames(ctx, fullListing.Categories.ID)
	if err != nil {
		return listing.FullListingResponse{}, err
	}

	resp := listing.FullListingResponse{
		ID:               fullListing.Listing.ID,
		Title:            fullListing.Listing.Title,
		Description:      fullListing.Listing.Description,
		Price:            fullListing.Listing.Price,
		Currency:         fullListing.Listing.Currency,
		OriginalPrice:    fullListing.Listing.Price,
		OriginalCurrency: fullListing.Listing.Currency,
		Location:         fullListing.Location,
		Categories:       categories,
		Characteristics:  make(map[string]interface{}),
		Images:           []string{},
		Boosts:           boosts,
		CreatedAt:        fullListing.Listing.CreatedAt.UnixMilli(),
		UpdatedAt:        fullListing.Listing.UpdatedAt.UnixMilli(),
	}

	for key, value := range fullListing.Characteristics {
		if value != nil {
			resp.Characteristics[key] = value
		}
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
	// Преобразуем структуру из конфига в формат для API
	configCategories := config.GetConfig().Categories.Data.Categories
	categories := make([]listing.CategoryNode, 0, len(configCategories))

	// Преобразуем формат категорий из TOML в формат API
	for _, cat := range configCategories {
		categories = append(categories, convertCategoryNodeToAPI(cat))
	}

	// Получаем язык из контекста
	lang := parser.GetLang(ctx)
	// Загружаем локализации
	var translations map[string]string
	var langData string

	switch lang {
	case models.LanguageRu:
		langData = config.GetConfig().Categories.LangCategories.Ru
	case models.LanguageEn:
		langData = config.GetConfig().Categories.LangCategories.En
	case models.LanguageEs:
		langData = config.GetConfig().Categories.LangCategories.Es
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

func (s *Listing) GetCharacteristicsForCategory(ctx context.Context, categoryIds []string) (listing.GetCharacteristicsForCategoryResponse, error) {
	// Получаем характеристики и переводы
	_, characteristicKeys, translations, err := s.getCategoryCharacteristics(ctx, categoryIds)
	if err != nil {
		return listing.GetCharacteristicsForCategoryResponse{}, err
	}

	// Получаем опции характеристик из конфига
	characteristicOptions := config.GetConfig().Categories.CategoryCharacteristics

	// Создаем массив характеристик в порядке их ключей
	result := make(listing.CharacteristicParam, 0, len(characteristicKeys))

	for _, key := range characteristicKeys {
		// Создаем параметр в зависимости от типа характеристики
		param := s.createParamForCharacteristic(key, characteristicOptions, translations)

		// Добавляем новый элемент напрямую в массив
		result = append(result, listing.CharacteristicParamItem{
			Role:  key,
			Param: param,
		})
	}

	return listing.GetCharacteristicsForCategoryResponse{
		CharacteristicParams: result,
	}, nil
}

// GetFiltersForCategory возвращает фильтры для указанной категории
func (s *Listing) GetFiltersForCategory(ctx context.Context, categoryId string) (listing.GetFiltersForCategoryResponse, error) {
	// Получаем значения характеристик из БД
	charValues, err := s.s.GetCategoryFilters(ctx, categoryId)
	if err != nil {
		return listing.GetFiltersForCategoryResponse{}, fmt.Errorf("error getting characteristic values: %w", err)
	}

	// Фильтруем пустые значения
	validFilters := s.filterEmptyValues(charValues)

	// Возвращаем только валидные фильтры
	return listing.GetFiltersForCategoryResponse{Filters: validFilters}, nil
}

// filterEmptyValues фильтрует пустые значения из фильтров
func (s *Listing) filterEmptyValues(filters models.Filters) models.Filters {
	result := make(models.Filters)

	for key, filter := range filters {
		switch f := filter.(type) {
		case models.PriceFilter:
			// Добавляем фильтр цены только если есть реальный диапазон цен и минимальная цена не равна максимальной
			if f.Min < f.Max && (f.Min > 0 || f.Max > 0) {
				result[key] = f
			}

		case models.ColorFilter:
			// Добавляем фильтр цвета только если есть значения
			if len(f.Options) > 0 {
				result[key] = f
			}

		case models.DropdownFilter:
			// Добавляем фильтр выпадающего списка только если есть значения
			if len(f) > 0 {
				result[key] = f
			}

		case models.CheckboxFilter:
			// Для чекбоксов проверяем, что значение не nil
			if f != nil {
				result[key] = f
			}

		case models.DimensionFilter:
			// Добавляем фильтр размеров только если есть реальные значения и корректная единица измерения
			if (f.Min > 0 || f.Max > 0) && f.Min <= f.Max && f.Dimension != "" {
				result[key] = f
			}

		default:
			// Для неизвестных типов фильтров просто копируем
			result[key] = filter
		}
	}

	return result
}

// getCategoryCharacteristics получает характеристики для указанных категорий
func (s *Listing) getCategoryCharacteristics(ctx context.Context, categoryIds []string) (map[string][]string, []string, map[string]string, error) {
	// Получаем язык из контекста
	lang := string(parser.GetLang(ctx))

	// Собираем характеристики категорий из структуры категорий в конфиге
	categoryCharacteristics := make(map[string][]string)

	// Собираем характеристики из категорий в конфиге
	for _, cat := range config.GetConfig().Categories.Data.Categories {
		charRoles := extractCharacteristicRoles(cat)
		categoryCharacteristics[cat.ID] = charRoles
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

// extractCharacteristicRoles извлекает роли характеристик из категории
func extractCharacteristicRoles(category config.CategoryNode) []string {
	roles := make([]string, 0, len(category.Characteristics))

	// Добавляем роли характеристик текущей категории
	for _, char := range category.Characteristics {
		roles = append(roles, char.Role)
	}

	return roles
}

// createParamForCharacteristic создает параметр нужного типа для характеристики
func (s *Listing) createParamForCharacteristic(characteristicKey string, options map[string][]string, translations map[string]string) interface{} {
	// Получаем тип параметра из мапы
	paramType, exists := models.CharacteristicParamMap[characteristicKey]
	if !exists {
		// Если тип не определен, возвращаем nil
		return nil
	}

	// Получаем язык из контекста
	lang := context.Background().Value(models.KeyLanguage)

	// Выбираем карту переводов в зависимости от языка
	var langOptions map[string]map[string]string
	switch lang {
	case string(models.LanguageRu):
		langOptions = config.GetConfig().Categories.OptionsTranslations.Ru
	case string(models.LanguageEn):
		langOptions = config.GetConfig().Categories.OptionsTranslations.En
	case string(models.LanguageEs):
		langOptions = config.GetConfig().Categories.OptionsTranslations.Es
	default:
		langOptions = config.GetConfig().Categories.OptionsTranslations.En
	}

	// В зависимости от типа параметра создаем соответствующую структуру
	switch paramType.(type) {
	case models.ColorParam:
		// Для цвета просто возвращаем пустую структуру
		return models.ColorParam{}

	case []models.DropdownOptionItem:
		// Для строковых параметров (выпадающих списков) загружаем опции
		paramOptions := []models.DropdownOptionItem{}

		// Получаем опции для данной характеристики
		if optionValues, ok := options[characteristicKey]; ok && len(optionValues) > 0 {
			for _, value := range optionValues {
				// По умолчанию используем значение как метку
				label := value

				// Если есть переводы для этой характеристики
				if optionTranslations, ok := langOptions[characteristicKey]; ok {
					// Если есть перевод для этого значения
					if translation, ok := optionTranslations[value]; ok {
						label = translation
					}
				}

				paramOptions = append(paramOptions, models.DropdownOptionItem{
					Value: value,
					Label: label,
				})
			}
		}

		return paramOptions

	case models.CheckboxParam:
		// Для чекбокса просто возвращаем пустую структуру
		return models.CheckboxParam{}

	case models.AmountParam:
		// Для числовых параметров добавляем соответствующую единицу измерения
		dimension := s.getDefaultDimensionForCharacteristic(characteristicKey)
		return models.AmountParam{
			DimensionOptions: dimension,
		}

	default:
		// Для неизвестных типов возвращаем nil
		return nil
	}
}

// getDefaultDimensionForCharacteristic возвращает дефолтную единицу измерения для числовой характеристики
func (s *Listing) getDefaultDimensionForCharacteristic(characteristicKey string) models.Dimension {
	switch characteristicKey {
	case models.CHAR_HEIGHT, models.CHAR_WIDTH, models.CHAR_DEPTH:
		// Для линейных размеров (дефолт - сантиметры)
		return models.Dimension(models.CM)

	case models.CHAR_AREA:
		// Для площади (дефолт - квадратные метры)
		return models.Dimension(models.M2)

	case models.CHAR_VOLUME:
		// Для объема (дефолт - литры)
		return models.Dimension(models.L)

	case models.CHAR_WEIGHT:
		// Для веса (дефолт - килограммы)
		return models.Dimension(models.KG)

	default:
		return models.Dimension("")
	}
}

// convertCategoryNodeToAPI преобразует структуру категории из конфига в формат API
func convertCategoryNodeToAPI(configNode config.CategoryNode) listing.CategoryNode {
	node := listing.CategoryNode{
		Category: listing.Category{
			ID: configNode.ID,
		},
	}

	// Преобразуем подкатегории рекурсивно
	if len(configNode.Subcategories) > 0 {
		node.Subcategories = make([]listing.CategoryNode, 0, len(configNode.Subcategories))
		for _, subCat := range configNode.Subcategories {
			node.Subcategories = append(node.Subcategories, convertCategoryNodeToAPI(subCat))
		}
	}

	return node
}
