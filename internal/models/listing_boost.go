package models

import (
	"github.com/google/uuid"
)

// ListingBoost представляет буст объявления в базе данных
type Boost struct {
	ListingID  uuid.UUID `json:"listing_id"`
	Type  BoostType `json:"type"`
	Commission float64   `json:"commission"`
}

// BoostType представляет типы бустов объявлений
type BoostType string

// Константы для типов бустов
const (
	BoostTypeBase   BoostType = "base"
	BoostTypeHighlight BoostType = "highlight"
	BoostTypeUpfront BoostType = "upfront"
)

// GetBoostTypesWithCommissions возвращает список всех доступных типов бустов с комиссиями
func GetBoostTypesWithCommissions() map[BoostType]float64 {
	return map[BoostType]float64{
		BoostTypeBase:   0.02,
		BoostTypeHighlight: 0.07,
		BoostTypeUpfront: 0.12,
	}
}

// String возвращает строковое представление типа буста
func (bt BoostType) String() string {
	return string(bt)
}
