package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Константы для типов фильтров
const (
	PRICE_TYPE     = "price"
	COLOR_TYPE     = "color"
	DROPDOWN_TYPE  = "dropdown"
	CHECKBOX_TYPE  = "checkbox"
	DIMENSION_TYPE = "dimension"
)

var RoleFilters = []string{
	PRICE_TYPE,
	COLOR_TYPE,
	DROPDOWN_TYPE,
	CHECKBOX_TYPE,
	DIMENSION_TYPE,
}

// ListingFilters представляет характеристики объявления в базе данных
type ListingFilters struct {
	ListingID uuid.UUID `json:"listing_id"`
	Filters   Filters   `json:"filters"`
}

// Структуры для различных типов фильтров
type PriceFilter struct {
	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type ColorFilter []string

type DropdownFilter []string

type CheckboxFilter bool

type DimensionFilter struct {
	Min       int    `json:"min"`
	Max       int    `json:"max"`
	Dimension string `json:"dimension"`
}

// Filters теперь map с ключами-строками и значениями-интерфейсами
type Filters map[string]interface{}

// FilterItem представляет элемент фильтра для API
type FilterItem struct {
	Role  string      `json:"role"`
	Type  string      `json:"type"`
	Param interface{} `json:"param"`
}

func (c Filters) GetPriceFilter(key string) (PriceFilter, bool) {
	var priceFilter PriceFilter
	value, ok := c[key]
	if !ok {
		return priceFilter, false
	}

	priceFilter, ok = value.(PriceFilter)
	if !ok {
		return priceFilter, false
	}

	return priceFilter, true
}

func (c Filters) GetColorFilter(key string) (ColorFilter, bool) {
	var colorFilter ColorFilter
	value, ok := c[key]
	if !ok {
		return colorFilter, false
	}

	colorFilter, ok = value.(ColorFilter)
	if !ok {
		return colorFilter, false
	}

	return colorFilter, true
}

func (c Filters) GetDropdownFilter(key string) (DropdownFilter, bool) {
	var dropdownFilter DropdownFilter
	value, ok := c[key]
	if !ok {
		return dropdownFilter, false
	}

	dropdownFilter, ok = value.(DropdownFilter)
	if !ok {
		return dropdownFilter, false
	}

	return dropdownFilter, true
}

func (c Filters) GetCheckboxFilter(key string) (CheckboxFilter, bool) {
	var checkboxFilter CheckboxFilter
	value, ok := c[key]
	if !ok {
		return checkboxFilter, false
	}

	checkboxFilter, ok = value.(CheckboxFilter)
	if !ok {
		return checkboxFilter, false
	}

	return checkboxFilter, true
}

func (c Filters) GetDimensionFilter(key string) (DimensionFilter, bool) {
	var dimensionFilter DimensionFilter
	value, ok := c[key]
	if !ok {
		return dimensionFilter, false
	}

	dimensionFilter, ok = value.(DimensionFilter)
	if !ok {
		return dimensionFilter, false
	}

	return dimensionFilter, true
}

func (c *Filters) UnmarshalJSON(data []byte) error {
	// Инициализируем map, если он nil
	if *c == nil {
		*c = make(Filters)
	}

	// Разбираем JSON как массив фильтров
	var filters []struct {
		Role  string          `json:"role"`
		Value json.RawMessage `json:"value"`
	}

	// Разбираем JSON как массив фильтров
	if err := json.Unmarshal(data, &filters); err != nil {
		return fmt.Errorf("failed to unmarshal JSON as array: %v", err)
	}

	// Обрабатываем фильтры
	for _, filter := range filters {
		switch filter.Role {
		case CHAR_PRICE:
			var priceFilter PriceFilter
			if err := json.Unmarshal(filter.Value, &priceFilter); err != nil {
				return fmt.Errorf("failed to parse price filter: %v", err)
			}
			(*c)[filter.Role] = priceFilter
		case CHAR_COLOR:
			var colorFilter ColorFilter
			if err := json.Unmarshal(filter.Value, &colorFilter); err != nil {
				return fmt.Errorf("failed to parse color filter: %v", err)
			}
			(*c)[filter.Role] = colorFilter
		case CHAR_BRAND, CHAR_CONDITION, CHAR_SEASON:
			var dropdownFilter DropdownFilter
			if err := json.Unmarshal(filter.Value, &dropdownFilter); err != nil {
				return fmt.Errorf("failed to parse dropdown filter: %v", err)
			}
			(*c)[filter.Role] = dropdownFilter
		case CHAR_STOCKED:
			var checkboxFilter CheckboxFilter
			if err := json.Unmarshal(filter.Value, &checkboxFilter); err != nil {
				return fmt.Errorf("failed to parse checkbox filter: %v", err)
			}
			(*c)[filter.Role] = checkboxFilter
		case CHAR_HEIGHT, CHAR_WIDTH, CHAR_DEPTH, CHAR_WEIGHT, CHAR_AREA, CHAR_VOLUME:
			var dimensionFilter DimensionFilter
			if err := json.Unmarshal(filter.Value, &dimensionFilter); err != nil {
				return fmt.Errorf("failed to parse dimension filter: %v", err)
			}
			(*c)[filter.Role] = dimensionFilter
		}
	}

	return nil
}

func (c Filters) MarshalJSON() ([]byte, error) {
	// Создаем тип для фильтра
	type Filter struct {
		Role  string      `json:"role"`
		Value interface{} `json:"value"`
	}

	// Создаем массив фильтров
	filters := []Filter{}

	// Добавляем все фильтры из map в массив
	for role, value := range c {
		filters = append(filters, Filter{
			Role:  role,
			Value: value,
		})
	}

	// Если нет фильтров, возвращаем пустой массив
	if len(filters) == 0 {
		return []byte("[]"), nil
	}

	// Сериализуем массив в JSON
	return json.Marshal(filters)
}
