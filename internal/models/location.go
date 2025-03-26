package models

import (
	"fmt"
	"github.com/google/uuid"
)

type Location struct {
	ID        string    `json:"id"`
	ListingID uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	Area      Area      `json:"area"`
}

// IsValid проверяет, является ли местоположение допустимым
func (l Location) IsValid() bool {
	// Проверяем, что имя не пустое
	if l.Name == "" {
		return false
	}
	// Проверяем валидность области
	return l.Area.IsValid()
}

// Validate проверяет валидность местоположения и возвращает ошибку, если местоположение недопустимо
func (l Location) Validate() error {
	if l.Name == "" {
		return fmt.Errorf("имя местоположения не может быть пустым")
	}
	
	if err := l.Area.Validate(); err != nil {
		return fmt.Errorf("недопустимая область: %w", err)
	}
	
	return nil
}

type Area struct {
	Coordinates struct {
		Lat float64 `json:"lat" validate:"required"`
		Lng float64 `json:"lng" validate:"required"`
	} `json:"coordinates" validate:"required"`
	Radius int `json:"radius" validate:"required"`
}

// IsValid проверяет, является ли область допустимой
func (a Area) IsValid() bool {
	// Проверяем, что координаты находятся в допустимом диапазоне
	if a.Coordinates.Lat < -90 || a.Coordinates.Lat > 90 {
		return false
	}
	if a.Coordinates.Lng < -180 || a.Coordinates.Lng > 180 {
		return false
	}
	// Проверяем, что радиус положительный
	if a.Radius <= 0 {
		return false
	}
	return true
}

// Validate проверяет валидность области и возвращает ошибку, если область недопустима
func (a Area) Validate() error {
	if a.Coordinates.Lat < -90 || a.Coordinates.Lat > 90 {
		return fmt.Errorf("широта должна быть в диапазоне от -90 до 90")
	}
	if a.Coordinates.Lng < -180 || a.Coordinates.Lng > 180 {
		return fmt.Errorf("долгота должна быть в диапазоне от -180 до 180")
	}
	if a.Radius <= 0 {
		return fmt.Errorf("радиус должен быть положительным числом")
	}
	return nil
}
