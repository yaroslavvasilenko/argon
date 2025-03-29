package service

import (
	"context"
	"math"
	"strconv"

	"github.com/yaroslavvasilenko/argon/internal/models"
)

// GetLocationById получает информацию о локации по её ID
func (l *Location) GetLocationById(ctx context.Context, locationID string) (models.Location, error) {
	// Получаем данные о локации из Nominatim с использованием нового метода хранилища
	locResp, err := l.s.GetLocationByID(ctx, locationID)
	if err != nil {
		// Если произошла ошибка, возвращаем пустую локацию
		return models.Location{}, err
	}

	// Вычисляем радиус на основе ограничивающего прямоугольника
	radius := calculateRadiusFromBoundingBox(locResp.BoundingBox, parseFloat(locResp.Lat), parseFloat(locResp.Lon))

	// Преобразуем ответ из Nominatim в модель локации
	location := models.Location{
		ID:   locationID,
		Name: locResp.DisplayName,
		Area: models.Area{
			Coordinates: struct {
				Lat float64 `json:"lat" validate:"required"`
				Lng float64 `json:"lng" validate:"required"`
			}{
				Lat: parseFloat(locResp.Lat),
				Lng: parseFloat(locResp.Lon),
			},
			Radius: radius,
		},
	}

	return location, nil
}

// parseFloat преобразует строку в float64, возвращая 0 в случае ошибки
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// calculateRadiusFromBoundingBox вычисляет приблизительный радиус на основе ограничивающего прямоугольника
func calculateRadiusFromBoundingBox(boundingBox []string, centerLat, centerLon float64) int {
	if len(boundingBox) < 4 {
		// Если нет данных о границах, используем стандартный радиус
		return 1000
	}

	// Парсим границы прямоугольника
	// boundingBox в Nominatim имеет формат [minLat, maxLat, minLon, maxLon]
	minLat := parseFloat(boundingBox[0])
	maxLat := parseFloat(boundingBox[1])
	minLon := parseFloat(boundingBox[2])
	maxLon := parseFloat(boundingBox[3])

	// Вычисляем расстояния от центра до углов прямоугольника
	distNE := haversineDistance(centerLat, centerLon, maxLat, maxLon)
	distNW := haversineDistance(centerLat, centerLon, maxLat, minLon)
	distSE := haversineDistance(centerLat, centerLon, minLat, maxLon)
	distSW := haversineDistance(centerLat, centerLon, minLat, minLon)

	// Берем максимальное расстояние как радиус
	radius := math.Max(math.Max(distNE, distNW), math.Max(distSE, distSW))

	// Округляем до целого числа метров
	return int(math.Round(radius))
}

// haversineDistance вычисляет расстояние между двумя точками на земле в метрах
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Радиус Земли в метрах
	const earthRadius = 6371000.0

	// Переводим градусы в радианы
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Разница координат
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Формула гаверсинуса
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}
