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

// Константы для вариантов сортировки
const (
	SORT_PRICE_ASC  = "price_asc"
	SORT_PRICE_DESC = "price_desc"
	SORT_NEWEST     = "newest"
	SORT_RELEVANCE  = "relevance"
)

var RoleFilters = []string{
	PRICE_TYPE,
	COLOR_TYPE,
	DROPDOWN_TYPE,
	CHECKBOX_TYPE,
	DIMENSION_TYPE,
}

// SortOrders содержит допустимые варианты сортировки
var SortOrders = []string{
	SORT_PRICE_ASC,
	SORT_PRICE_DESC,
	SORT_NEWEST,
	SORT_RELEVANCE,
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

type CheckboxFilter *bool

type DimensionFilter struct {
	Min       int    `json:"min"`
	Max       int    `json:"max"`
	Dimension string `json:"dimension"`
}

// Filters теперь map с ключами-строками и значениями-интерфейсами
type Filters map[string]interface{}

// Validate проверяет валидность фильтров и возвращает ошибку, если фильтры недопустимы
func (f Filters) Validate() error {
	for key, value := range f {
		if key == "" {
			return fmt.Errorf("filter key cannot be empty")
		}
		
		switch key {
		case CHAR_PRICE:
			_, ok := value.(PriceFilter)
			if !ok {
				_, ok := value.(map[string]interface{})
				if !ok {
					return fmt.Errorf("filter '%s' must be of type PriceFilter", key)
				}
			}
		case CHAR_COLOR:
			// Проверяем, что значение - это ColorFilter или может быть преобразовано в него
			_, ok := value.(ColorFilter)
			if !ok {
				// Попробуем преобразовать из строки или массива строк
				_, ok := value.(string)
				if !ok {
					_, ok := value.([]string)
					if !ok {
						return fmt.Errorf("filter '%s' must be of type ColorFilter, string or []string", key)
					}
				}
			}
		case CHAR_BRAND, CHAR_CONDITION, CHAR_SEASON:
			// Проверяем, что значение - это DropdownFilter или может быть преобразовано в него
			_, ok := value.(DropdownFilter)
			if !ok {
				// Попробуем преобразовать из строки или массива строк
				_, ok := value.(string)
				if !ok {
					_, ok := value.([]string)
					if !ok {
						return fmt.Errorf("фильтр '%s' должен быть типа DropdownFilter, string или []string", key)
					}
				}
			}
		case CHAR_STOCKED:
			// Проверяем, что значение - это CheckboxFilter или может быть преобразовано в него
			_, ok := value.(CheckboxFilter)
			if !ok {
				// Попробуем преобразовать из булевого значения
				_, ok := value.(bool)
				if !ok {
					return fmt.Errorf("фильтр '%s' должен быть типа CheckboxFilter или bool", key)
				}
			}
		case CHAR_HEIGHT, CHAR_WIDTH, CHAR_DEPTH, CHAR_WEIGHT, CHAR_AREA, CHAR_VOLUME:
			// Проверяем, что значение - это DimensionFilter или может быть преобразовано в него
			dimFilter, ok := value.(DimensionFilter)
			if !ok {
				// Попробуем преобразовать из map
				_, ok := value.(map[string]interface{})
				if !ok {
					return fmt.Errorf("фильтр '%s' должен быть типа DimensionFilter", key)
				}
			} else {
				// Проверяем значения DimensionFilter
				if dimFilter.Min < 0 {
					return fmt.Errorf("фильтр '%s': минимальное значение не может быть отрицательным", key)
				}
				if dimFilter.Max < dimFilter.Min && dimFilter.Max != 0 {
					return fmt.Errorf("фильтр '%s': максимальное значение не может быть меньше минимального", key)
				}
				
				// Проверяем единицу измерения
				if dimFilter.Dimension == "" {
					return fmt.Errorf("фильтр '%s': единица измерения не может быть пустой", key)
				}
				
				// Проверяем, что единица измерения допустима для данного типа характеристики
				if !isValidDimensionUnit(key, dimFilter.Dimension) {
					return fmt.Errorf("фильтр '%s': недопустимая единица измерения '%s'", key, dimFilter.Dimension)
				}
			}
		}
	}
	
	return nil
}

// isValidDimensionUnit проверяет, является ли единица измерения допустимой для указанного типа характеристики
func isValidDimensionUnit(characteristicType string, unit string) bool {
	switch characteristicType {
	case CHAR_HEIGHT, CHAR_WIDTH, CHAR_DEPTH:
		// Для линейных размеров допустимы см, м, км
		return unit == CM || unit == M || unit == KM
	case CHAR_WEIGHT:
		// Для веса допустимы г, кг, т
		return unit == G || unit == KG || unit == T
	case CHAR_AREA:
		// Для площади допустимы см², м², км²
		return unit == CM2 || unit == M2 || unit == KM2
	case CHAR_VOLUME:
		// Для объема допустимы см³, м³, км³, мл, л
		return unit == CM3 || unit == M3 || unit == KM3 || unit == ML || unit == L
	default:
		return false
	}
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
		// Пробуем преобразовать из map
		mapValue, ok := value.(map[string]interface{})
		if !ok {
			return dimensionFilter, false
		}
		
		if min, ok := mapValue["min"].(float64); ok {
			dimensionFilter.Min = int(min)
		}
		if max, ok := mapValue["max"].(float64); ok {
			dimensionFilter.Max = int(max)
		}
		if dimension, ok := mapValue["dimension"].(string); ok {
			dimensionFilter.Dimension = dimension
		}
		
		return dimensionFilter, true
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
		Param json.RawMessage `json:"param"`
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
			if err := json.Unmarshal(filter.Param, &priceFilter); err != nil {
				return fmt.Errorf("failed to parse price filter: %v", err)
			}
			(*c)[filter.Role] = priceFilter
		case CHAR_COLOR:
			var colorFilter ColorFilter
			if err := json.Unmarshal(filter.Param, &colorFilter); err != nil {
				return fmt.Errorf("failed to parse color filter: %v", err)
			}
			(*c)[filter.Role] = colorFilter
		case CHAR_BRAND, CHAR_CONDITION, CHAR_SEASON:
			var dropdownFilter DropdownFilter
			if err := json.Unmarshal(filter.Param, &dropdownFilter); err != nil {
				return fmt.Errorf("failed to parse dropdown filter: %v", err)
			}
			(*c)[filter.Role] = dropdownFilter
		case CHAR_STOCKED:
			var checkboxFilter CheckboxFilter
			if err := json.Unmarshal(filter.Param, &checkboxFilter); err != nil {
				return fmt.Errorf("failed to parse checkbox filter: %v", err)
			}
			(*c)[filter.Role] = checkboxFilter
		case CHAR_HEIGHT, CHAR_WIDTH, CHAR_DEPTH, CHAR_WEIGHT, CHAR_AREA, CHAR_VOLUME:
			var dimensionFilter DimensionFilter
			if err := json.Unmarshal(filter.Param, &dimensionFilter); err != nil {
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
		Param interface{} `json:"param"`
	}

	// Создаем массив фильтров
	filters := []Filter{}

	// Добавляем все фильтры из map в массив
	for role, value := range c {
		filters = append(filters, Filter{
			Role:  role,
			Param: value,
		})
	}

	// Если нет фильтров, возвращаем пустой массив
	if len(filters) == 0 {
		return []byte("[]"), nil
	}

	// Сериализуем массив в JSON
	return json.Marshal(filters)
}
