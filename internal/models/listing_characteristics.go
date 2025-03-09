package models

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	PRICE_FILTER     = "price"
	COLOR_FILTER     = "color"
	DROPDOWN_FILTER  = "dropdown"
	SELECTOR_FILTER  = "selector"
	CHECKBOX_FILTER  = "checkbox"
	DIMENSION_FILTER = "dimension"
)

var RoleFilters = []string{
	PRICE_FILTER,
	COLOR_FILTER,
	DROPDOWN_FILTER,
	SELECTOR_FILTER,
	CHECKBOX_FILTER,
	DIMENSION_FILTER,
}

// ListingCharacteristics представляет характеристики объявления в базе данных
type ListingCharacteristics struct {
	ListingID       uuid.UUID       `json:"listing_id"`
	Characteristics Characteristics `json:"characteristics"`
}

// Структуры для различных типов фильтров
type PriceFilter struct {
	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type ColorFilter []string

type DropdownFilter []string

type SelectorFilter string

type CheckboxFilter bool

type DimensionFilter struct {
	Min       int    `json:"min,omitempty"`
	Max       int    `json:"max,omitempty"`
	Dimension string `json:"dimension,omitempty"`
}

// Characteristics теперь map с ключами-строками и значениями-интерфейсами
type Characteristics map[string]interface{}

func (c *Characteristics) GetPriceFilter() (PriceFilter, bool) {
	var priceFilter PriceFilter
	value, ok := (*c)[PRICE_FILTER]
	if !ok {
		return priceFilter, false
	}

	priceFilter, ok = value.(PriceFilter)
	if !ok {
		return priceFilter, false
	}

	return priceFilter, true
}

func (c *Characteristics) GetColorFilter() (ColorFilter, bool) {
	var colorFilter ColorFilter
	value, ok := (*c)[COLOR_FILTER]
	if !ok {
		return colorFilter, false
	}

	colorFilter, ok = value.(ColorFilter)
	if !ok {
		return colorFilter, false
	}

	return colorFilter, true
}

func (c *Characteristics) GetDropdownFilter() (DropdownFilter, bool) {
	var dropdownFilter DropdownFilter
	value, ok := (*c)[DROPDOWN_FILTER]
	if !ok {
		return dropdownFilter, false
	}

	dropdownFilter, ok = value.(DropdownFilter)
	if !ok {
		return dropdownFilter, false
	}

	return dropdownFilter, true
}

func (c *Characteristics) GetSelectorFilter() (SelectorFilter, bool) {
	var selectorFilter SelectorFilter
	value, ok := (*c)[SELECTOR_FILTER]
	if !ok {
		return selectorFilter, false
	}

	selectorFilter, ok = value.(SelectorFilter)
	if !ok {
		return selectorFilter, false
	}

	return selectorFilter, true
}

func (c *Characteristics) GetCheckboxFilter() (CheckboxFilter, bool) {
	var checkboxFilter CheckboxFilter
	value, ok := (*c)[CHECKBOX_FILTER]
	if !ok {
		return checkboxFilter, false
	}

	checkboxFilter, ok = value.(CheckboxFilter)
	if !ok {
		return checkboxFilter, false
	}

	return checkboxFilter, true
}

func (c *Characteristics) GetDimensionFilter() (DimensionFilter, bool) {
	var dimensionFilter DimensionFilter
	value, ok := (*c)[DIMENSION_FILTER]
	if !ok {
		return dimensionFilter, false
	}

	dimensionFilter, ok = value.(DimensionFilter)
	if !ok {
		return dimensionFilter, false
	}

	return dimensionFilter, true
}

func (c *Characteristics) UnmarshalJSON(data []byte) error {
	// Инициализируем map, если он nil
	if *c == nil {
		*c = make(Characteristics)
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
		case PRICE_FILTER:
			var priceFilter PriceFilter
			if err := json.Unmarshal(filter.Value, &priceFilter); err != nil {
				return fmt.Errorf("failed to parse price filter: %v", err)
			}
			(*c)[PRICE_FILTER] = priceFilter
		case COLOR_FILTER:
			var colorFilter ColorFilter
			if err := json.Unmarshal(filter.Value, &colorFilter); err != nil {
				return fmt.Errorf("failed to parse color filter: %v", err)
			}
			(*c)[COLOR_FILTER] = colorFilter
		case DROPDOWN_FILTER:
			var dropdownFilter DropdownFilter
			if err := json.Unmarshal(filter.Value, &dropdownFilter); err != nil {
				return fmt.Errorf("failed to parse dropdown filter: %v", err)
			}
			(*c)[DROPDOWN_FILTER] = dropdownFilter
		case SELECTOR_FILTER:
			var selectorFilter SelectorFilter
			if err := json.Unmarshal(filter.Value, &selectorFilter); err != nil {
				return fmt.Errorf("failed to parse selector filter: %v", err)
			}
			(*c)[SELECTOR_FILTER] = selectorFilter
		case CHECKBOX_FILTER:
			var checkboxFilter CheckboxFilter
			if err := json.Unmarshal(filter.Value, &checkboxFilter); err != nil {
				return fmt.Errorf("failed to parse checkbox filter: %v", err)
			}
			(*c)[CHECKBOX_FILTER] = checkboxFilter
		case DIMENSION_FILTER:
			var dimensionFilter DimensionFilter
			if err := json.Unmarshal(filter.Value, &dimensionFilter); err != nil {
				return fmt.Errorf("failed to parse dimension filter: %v", err)
			}
			(*c)[DIMENSION_FILTER] = dimensionFilter
		default:
			// Для неизвестных типов фильтров просто сохраняем значение как есть
			var value interface{}
			if err := json.Unmarshal(filter.Value, &value); err != nil {
				return fmt.Errorf("failed to parse unknown filter %s: %v", filter.Role, err)
			}
			(*c)[filter.Role] = value
		}
	}

	return nil
}

func (c Characteristics) MarshalJSON() ([]byte, error) {
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
