package storage

import (
	"context"
	"encoding/json"

	"github.com/yaroslavvasilenko/argon/internal/models"
)

// getListingWithRelatedData получает полные данные объявления, включая категории, бусты, характеристики и местоположение
func (s *Listing) getListingWithRelatedData(ctx context.Context, listing models.Listing) (models.ListingResult, error) {
	// Создаем результат с основными данными объявления
	result := models.NewListingResult(listing)

	// Получаем категории объявления
	// Сначала получаем идентификаторы категорий как строки
	var categoryIDs []string
	if err := s.gorm.Table("listing_categories").
		Select("category_id").
		Where("listing_id = ?", listing.ID).
		Pluck("category_id", &categoryIDs).Error; err != nil {
		return models.ListingResult{}, err
	}

	// Затем создаем объекты категорий с правильной структурой
	if len(categoryIDs) > 0 {
		// Создаем одну категорию с массивом идентификаторов
		category := models.Category{
			ID:        categoryIDs,
			ListingID: listing.ID.String(),
		}
		result.SetCategories([]models.Category{category})
	}

	// Получаем бусты объявления
	var boosts []models.Boost
	if err := s.gorm.Table("listing_boosts").
		Select("listing_id, boost_type, commission").
		Where("listing_id = ?", listing.ID).
		Find(&boosts).Error; err != nil {
		return models.ListingResult{}, err
	}
	result.SetBoosts(boosts)

	// Получаем характеристики объявления
	var characteristicsBytes []byte
	if err := s.gorm.Table("listing_characteristics").
		Select("characteristics").
		Where("listing_id = ?", listing.ID).
		Row().Scan(&characteristicsBytes); err != nil {
		// Если характеристики не найдены, просто продолжаем без них
		// Это не критическая ошибка
	} else if len(characteristicsBytes) > 0 {
		// Преобразуем JSON-байты в map
		var characteristics map[string]interface{}
		if err := json.Unmarshal(characteristicsBytes, &characteristics); err != nil {
			return models.ListingResult{}, err
		}
		result.SetCharacteristics(characteristics)
	}

	// Получаем местоположение объявления
	// Создаем переменные для хранения координат и радиуса
	var latitude, longitude float64
	var radius int

	// Получаем данные из БД напрямую в локальные переменные
	var location models.Location
	if err := s.gorm.Table("locations").
		Select("id, listing_id, name, latitude, longitude, radius").
		Where("listing_id = ?", listing.ID).
		Row().Scan(&location.ID, &location.ListingID, &location.Name, &latitude, &longitude, &radius); err != nil {
		// Если местоположение не найдено, просто продолжаем без него
		// Это не критическая ошибка
	} else {
		// Преобразуем данные из БД в структуру Area
		location.Area = models.Area{
			Coordinates: models.Coordinates{
				Lat: latitude,
				Lng: longitude,
			},
			Radius: radius,
		}
		result.SetLocation(location)
	}

	return result, nil
}
