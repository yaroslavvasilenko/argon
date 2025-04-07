package models

import (
	"encoding/json"
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

// ColorParam представляет параметр цвета
type ColorParam struct {
	// Пустая структура, так как ограничения нужны только для фильтров
}

// StringParam представляет параметр выпадающего списка
type DropdownOptionItem struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

// Dimension представляет единицу измерения
type Dimension string

// AmountParam представляет параметр с числовым значением и единицей измерения
type AmountParam struct {
	Value            float64    `json:"value" validate:"required"`
	DimensionOptions Dimension `json:"dimension_options" validate:"required"`
}

// CheckboxParam представляет параметр чекбокса
type CheckboxParam struct {
	// Пустая структура, так как для чекбокса не требуются ограничительные параметры
}

// CharacteristicParam представляет интерфейс для всех типов параметров
type CharacteristicParam interface{}

// CharacteristicParamMap сопоставляет характеристики с их типами параметров
var CharacteristicParamMap = map[string]interface{}{
	// Цвет
	CHAR_COLOR: ColorParam{},

	// Выпадающие списки
	CHAR_CONDITION: DropdownOptionItem{},
	CHAR_SEASON:    DropdownOptionItem{},
	CHAR_BRAND:     DropdownOptionItem{},

	// Чекбоксы
	CHAR_STOCKED: CheckboxParam{},

	// Размеры и измерения
	CHAR_HEIGHT: AmountParam{},
	CHAR_WIDTH:  AmountParam{},
	CHAR_DEPTH:  AmountParam{},
	CHAR_WEIGHT: AmountParam{},
	CHAR_AREA:   AmountParam{},
	CHAR_VOLUME: AmountParam{},
}

// Characteristic представляет собой карту характеристик, где ключ - это роль, а значение - это значение характеристики
type Characteristic map[string]interface{}

// CharacteristicItem представляет отдельную характеристику с ролью и значением
type CharacteristicItem struct {
	Role  string      `json:"role"`
	Value interface{} `json:"value"`
}

func CreateCharacteristics(keys []string, translations map[string]string) Characteristic {
	result := make(Characteristic, len(keys))

	for _, key := range keys {
		// Получаем перевод для ключа
		translation, ok := translations[key]
		if !ok {
			translation = key // Если перевод не найден, используем ключ
		}

		// Создаем характеристику с ключом и переводом
		result[key] = translation
	}

	return result
}

// MarshalJSON реализует интерфейс json.Marshaler для типа Characteristic
func (c Characteristic) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	charItems := make([]CharacteristicItem, 0, len(c))
	for role, param := range c {
		charItems = append(charItems, CharacteristicItem{
			Role:  role,
			Value: param,
		})
	}

	return json.Marshal(charItems)
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для типа Characteristic
func (c *Characteristic) UnmarshalJSON(data []byte) error {
	var charItems []CharacteristicItem
	if err := json.Unmarshal(data, &charItems); err != nil {
		return err
	}

	if *c == nil {
		*c = make(Characteristic)
	}

	for _, item := range charItems {
		// Определяем тип параметра для данной характеристики
		paramType, ok := CharacteristicParamMap[item.Role]
		if !ok {
			continue
		}

		// Обрабатываем значение в зависимости от типа характеристики
		switch paramType.(type) {
		case ColorParam:
			// Для цвета ожидаем строку
			if strValue, ok := item.Value.(string); ok {
				(*c)[item.Role] = strValue
			}

		case DropdownOptionItem:
			// Для выпадающих списков ожидаем объект {value, label}
			switch v := item.Value.(type) {
			case map[string]interface{}:
				// Если это объект, преобразуем его в DropdownOptionItem
				option := DropdownOptionItem{}
				if value, ok := v["value"].(string); ok {
					option.Value = value
				}
				if label, ok := v["label"].(string); ok {
					option.Label = label
				}
				(*c)[item.Role] = option
			}

		case CheckboxParam:
			// Для чекбокса ожидаем булево значение
			if boolValue, ok := item.Value.(bool); ok {
				(*c)[item.Role] = boolValue
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
					}
				}
				
				// Обрабатываем поле dimension
				if dimensionField, ok := v["dimension"]; ok {
					if dimension, ok := dimensionField.(string); ok {
						amountParam.DimensionOptions = Dimension(dimension)
					}
				}
				
				(*c)[item.Role] = amountParam
			}
		}
	}

	return nil
}
