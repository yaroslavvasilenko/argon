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
	Color string `json:"color"`
}

type DropdownOption struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

type CheckboxValue struct {
	CheckboxValue bool `json:"checkbox_value"`
}

type Amount struct {
	Value     float64   `json:"value" validate:"required"`
	Dimension Dimension `json:"dimension" validate:"required"`
}

// CharacteristicValue представляет собой карту характеристик, где ключ - это роль, а значение - это значение характеристики
type CharacteristicValue map[string]interface{}

// SetColor устанавливает значение цвета для указанной роли
func (c CharacteristicValue) SetColor(role string, value Color) {
    c[role] = value
}

// SetDropdownOption устанавливает значение выпадающего списка для указанной роли
func (c CharacteristicValue) SetDropdownOption(role string, value DropdownOption) {
    c[role] = value
}

// SetCheckboxValue устанавливает значение чекбокса для указанной роли
func (c CharacteristicValue) SetCheckboxValue(role string, value bool) {
    c[role] = value
}

// SetAmount устанавливает числовое значение с единицей измерения для указанной роли
func (c CharacteristicValue) SetAmount(role string, value Amount) {
    c[role] = value
}

// GetColor возвращает значение цвета для указанной роли
func (c CharacteristicValue) GetColor(role string) (Color, bool) {
    if v, ok := c[role].(Color); ok {
        return v, true
    }
    return Color{}, false
}

// GetDropdownOption возвращает значение выпадающего списка для указанной роли
func (c CharacteristicValue) GetDropdownOption(role string) (DropdownOption, bool) {
    if v, ok := c[role].(DropdownOption); ok {
        return v, true
    }
    return DropdownOption{}, false
}

// GetCheckboxValue возвращает значение чекбокса для указанной роли
func (c CharacteristicValue) GetCheckboxValue(role string) (bool, bool) {
    if v, ok := c[role].(bool); ok {
        return v, true
    }
    return false, false
}

// GetAmount возвращает числовое значение с единицей измерения для указанной роли
func (c CharacteristicValue) GetAmount(role string) (Amount, bool) {
    if v, ok := c[role].(Amount); ok {
        return v, true
    }
    return Amount{}, false
}

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
		paramTypeObj, ok := CharacteristicValueMap[item.Role]
		if !ok {
			continue
		}

		// Переменная для хранения обработанного значения
		var processedValue interface{}

		// Обрабатываем значение в зависимости от типа характеристики
		switch paramTypeObj.(type) {
		case Color:
			// Для цвета ожидаем строку с кодом цвета или объект с полем color
			switch v := item.Value.(type) {
			case string:
				processedValue = Color{Color: v}
			case map[string]interface{}:
				if colorStr, ok := v["color"].(string); ok {
					processedValue = Color{Color: colorStr}
				}
			}

		case DropdownOption:
			// Для выпадающих списков ожидаем объект с массивом options
			switch v := item.Value.(type) {
			case map[string]interface{}:
				// Если получили объект, проверяем наличие полей value и label
				option := DropdownOption{}
				if value, ok := v["value"].(string); ok {
					option.Value = value
				}
				if label, ok := v["label"].(string); ok {
					option.Label = label
				}
				processedValue = option
			case []interface{}:
				// Если получили массив, обрабатываем его как массив опций
				options := make([]DropdownOption, 0, len(v))
				for _, opt := range v {
					if optMap, ok := opt.(map[string]interface{}); ok {
						option := DropdownOption{}
						if value, ok := optMap["value"].(string); ok {
							option.Value = value
						}
						if label, ok := optMap["label"].(string); ok {
							option.Label = label
						}
						options = append(options, option)
					} else if strVal, ok := opt.(string); ok {
						// Если получили строку, используем её как value
						options = append(options, DropdownOption{Value: strVal})
					}
				}
				// Создаем DropdownOption из первой опции, если она есть
				if len(options) > 0 {
					processedValue = options[0]
				} else {
					processedValue = DropdownOption{}
				}
			}



		case CheckboxValue:
			// Для чекбокса ожидаем булево значение
			if checkboxValue, ok := item.Value.(map[string]interface{}); ok {
				checkboxValueParam := CheckboxValue{}
				if checkboxValue["checkbox_value"].(bool) {
					checkboxValueParam.CheckboxValue = true
				}
				processedValue = checkboxValueParam
			}

		case Amount:
			// Для AmountParam ожидаем объект {value, dimension}
			switch v := item.Value.(type) {
			case map[string]interface{}:
				amountParam := Amount{}

				// Обрабатываем поле value
				if valueField, ok := v["value"]; ok {
					switch vf := valueField.(type) {
					case float64:
						amountParam.Value = vf
					case int:
						amountParam.Value = float64(vf)
					case int64:
						amountParam.Value = float64(vf)
					case json.Number:
						if floatVal, err := vf.Float64(); err == nil {
							amountParam.Value = floatVal
						}
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
			case float64:
				// Если получили просто число, создаем AmountParam с дефолтной единицей измерения
				amountParam := Amount{Value: v}
				
				processedValue = amountParam
			case int, int64:
				// Если получили целое число, преобразуем в float64
				var floatVal float64
				switch val := v.(type) {
				case int:
					floatVal = float64(val)
				case int64:
					floatVal = float64(val)
				}
				
				amountParam := Amount{Value: floatVal}
				
				processedValue = amountParam
			}
		}

		// Добавляем обработанное значение в карту
		(*c)[item.Role] = processedValue
	}

	return nil
}
