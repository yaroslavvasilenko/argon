package models

import (
	"encoding/json"
	"strconv"
)

// Константы для типов характеристик
const (
	CHAR_PRICE = "price"

	CHAR_COLOR = "color"

	CHAR_CONDITION = "condition"

	CHAR_SEASON = "season"

	CHAR_BRAND = "brand"

	CHAR_STOCKED = "stocked"

	CHAR_HEIGHT = "height"
	CHAR_WIDTH  = "width"
	CHAR_DEPTH  = "depth"
	CHAR_WEIGHT = "weight"
	CHAR_AREA   = "area"
	CHAR_VOLUME = "volume"
)

// Константы единиц измерений для различных физических величин
const (
	// Длина
	CM = "cm" // сантиметр
	M  = "m"  // метр
	KM = "km" // километр

	// Площадь
	CM2 = "cm2" // квадратный сантиметр
	M2  = "m2"  // квадратный метр
	KM2 = "km2" // квадратный километр

	// Объем
	CM3 = "cm3" // кубический сантиметр
	M3  = "m3"  // кубический метр
	KM3 = "km3" // кубический километр
	ML  = "ml"  // миллилитр
	L   = "l"   // литр

	// Масса
	G  = "g"  // грамм
	KG = "kg" // килограмм
	T  = "t"  // тонна

	// Электричество
	MA = "ma" // миллиампер
	A  = "a"  // ампер
	W  = "w"  // ватт
	KW = "kw" // киловатт
	OM = "om" // ом
)

// CharacteristicValueMap сопоставляет характеристики с их типами параметров
var CharacteristicValueMap = map[string]interface{}{
	// Цвет
	CHAR_COLOR: Color{},

	// Выпадающие списки
	CHAR_CONDITION: DropdownOption{},
	CHAR_SEASON:    DropdownOption{},
	CHAR_BRAND:     DropdownOption{},

	// Чекбоксы
	CHAR_STOCKED: CheckboxValue{},

	// Размеры и измерения
	CHAR_HEIGHT: Amount{},
	CHAR_WIDTH:  Amount{},
	CHAR_DEPTH:  Amount{},
	CHAR_WEIGHT: Amount{},
	CHAR_AREA:   Amount{},
	CHAR_VOLUME: Amount{},
}

type Color struct {
	// Пустая структура, так как ограничения нужны только для фильтров
}

type DropdownOption struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

type CheckboxValue struct {
	// Пустая структура, так как для чекбокса не требуются ограничительные параметры
}

type Amount struct {
	Value            float64     `json:"value" validate:"required"`
	DimensionOptions []Dimension `json:"dimension_options" validate:"required"`
}

// CharacteristicValue представляет собой карту характеристик, где ключ - это роль, а значение - это значение характеристики
type CharacteristicValue map[string]interface{}

// Dimension представляет единицу измерения
type Dimension string

type CharacteristicRoleJSON struct {
	Role  string      `json:"role"`
	Value interface{} `json:"value"`
}

// MarshalJSON реализует интерфейс json.Marshaler для типа Characteristic
func (c CharacteristicValue) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	charItems := make([]CharacteristicRoleJSON, 0, len(c))
	for role, value := range c {
		charItems = append(charItems, CharacteristicRoleJSON{
			Role:  role,
			Value: value,
		})
	}

	return json.Marshal(charItems)
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для типа Characteristic
func (c *CharacteristicValue) UnmarshalJSON(data []byte) error {
	var charItems []CharacteristicRoleJSON
	if err := json.Unmarshal(data, &charItems); err != nil {
		return err
	}

	if *c == nil {
		*c = make(CharacteristicValue)
	}

	for _, item := range charItems {
		// Определяем тип параметра для данной характеристики
		paramType, ok := CharacteristicParamMap[item.Role]
		if !ok {
			continue
		}

		// Переменная для хранения обработанного значения
		var processedValue interface{}

		// Обрабатываем значение в зависимости от типа характеристики
		switch paramType.(type) {
		case ColorParam:
			// Для цвета ожидаем строку
			if strValue, ok := item.Value.(string); ok {
				processedValue = strValue
			}

		case StringParam:
			// Для выпадающих списков ожидаем объект с массивом options
			v := item.Value.(map[string]interface{})

			DropdownOption := DropdownOption{}
			if value, ok := v["value"].(string); ok {
				DropdownOption.Value = value
			}
			if label, ok := v["label"].(string); ok {
				DropdownOption.Label = label
			}
			processedValue = DropdownOption

		case CheckboxParam:
			// Для чекбокса ожидаем булево значение
			if boolValue, ok := item.Value.(bool); ok {
				processedValue = boolValue
			}

		case AmountParam:
			// Для AmountParam ожидаем объект {value, dimension}
			switch v := item.Value.(type) {
			case map[string]interface{}:
				amountParam := AmountParam{}

				// Обрабатываем поле value
				if valueField, ok := v["value"]; ok {
					switch vf := valueField.(type) {
					case float64:
						amountParam.Value = vf
					case string:
						// Пробуем преобразовать строку в float64
						if floatVal, err := strconv.ParseFloat(vf, 64); err == nil {
							amountParam.Value = floatVal
						}
					}
				}

				// Обрабатываем поле dimension
				if dimensionField, ok := v["dimension"]; ok {
					if dimension, ok := dimensionField.(string); ok {
						amountParam.Dimension = Dimension(dimension)
					}
				}

				processedValue = amountParam
			}
		}

		// Добавляем обработанное значение в карту
		(*c)[item.Role] = processedValue
	}

	return nil
}
