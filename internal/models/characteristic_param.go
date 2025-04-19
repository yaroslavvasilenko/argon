package models

import (
	"encoding/json"
)

// ColorParam представляет параметр цвета
type ColorParam struct {
	Color Color `json:"color"`
}

// StringParam представляет параметр выпадающего списка
type StringParam struct {
	Options []DropdownOption `json:"options"`
}

// CheckboxParam представляет параметр чекбокса
type CheckboxParam struct {
	// Пустая структура, так как для чекбокса не требуются ограничительные параметры
}

// AmountParam представляет параметр с числовым значением и единицей измерения
type AmountParam struct {
	Value     float64   `json:"value" validate:"required"`
	Dimension Dimension `json:"dimension" validate:"required"`
}

// CharacteristicParamItem представляет отдельную характеристику с ролью и параметром
// Используется только для сериализации и десериализации
type CharacteristicParamItem struct {
	Role  string      `json:"role"`
	Param interface{} `json:"param"`
}

// CharacteristicParam представляет собой карту параметров характеристик
type CharacteristicParam map[string]interface{}



// CharacteristicParamMap сопоставляет характеристики с их типами параметров
var CharacteristicParamMap = map[string]interface{}{
	// Цвет
	CHAR_COLOR: ColorParam{},

	// Выпадающие списки
	CHAR_CONDITION: StringParam{},
	CHAR_SEASON:    StringParam{},
	CHAR_BRAND:     StringParam{},

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



// MarshalJSON реализует интерфейс json.Marshaler для типа CharacteristicParam
func (c CharacteristicParam) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	charItems := make([]CharacteristicParamItem, 0, len(c))
	for role, param := range c {
		charItems = append(charItems, CharacteristicParamItem{
			Role:  role,
			Param: param,
		})
	}

	return json.Marshal(charItems)
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для типа CharacteristicParam
// func (c *CharacteristicParam) UnmarshalJSON(data []byte) error {
// 	var charItems []CharacteristicParamItem
// 	if err := json.Unmarshal(data, &charItems); err != nil {
// 		return err
// 	}

// 	if *c == nil {
// 		*c = make(CharacteristicParam)
// 	}

// 	for _, item := range charItems {
// 		// Определяем тип параметра для данной характеристики
// 		paramType, ok := CharacteristicParamMap[item.Role]
// 		if !ok {
// 			continue
// 		}

// 		// Переменная для хранения обработанного значения
// 		var processedParam interface{}

// 		// Обрабатываем параметр в зависимости от типа характеристики
// 		switch paramType.(type) {
// 		case ColorParam:
// 			// Для цвета ожидаем объект с полем color
// 			if mapParam, ok := item.Param.(map[string]interface{}); ok {
// 				if colorValue, ok := mapParam["color"]; ok {
// 					processedParam = colorValue
// 				}
// 			}

// 		case StringParam:
// 			// Для выпадающих списков ожидаем объект с массивом options
// 			if mapParam, ok := item.Param.(map[string]interface{}); ok {
// 				if optionsValue, ok := mapParam["options"]; ok {
// 					// Создаем StringParam с опциями
// 					stringParam := StringParam{
// 						Options: make([]DropdownOption, 0),
// 					}

// 					// Обрабатываем опции в зависимости от их типа
// 					switch opts := optionsValue.(type) {
// 					case []interface{}:
// 						// Обрабатываем массив интерфейсов
// 						for _, opt := range opts {
// 							switch o := opt.(type) {
// 							case map[string]interface{}:
// 								// Если это объект, извлекаем value и label
// 								option := DropdownOption{}
// 								if value, ok := o["value"].(string); ok {
// 									option.Value = value
// 								}
// 								if label, ok := o["label"].(string); ok {
// 									option.Label = label
// 								}
// 								stringParam.Options = append(stringParam.Options, option)
// 							case string:
// 								// Если это строка, используем её как value
// 								stringParam.Options = append(stringParam.Options, DropdownOption{Value: o})
// 							}
// 						}
// 					}

// 					processedParam = stringParam
// 				}
// 			}

// 		case CheckboxParam:
// 			// Для чекбокса ожидаем булево значение
// 			if mapParam, ok := item.Param.(map[string]interface{}); ok {
// 				if checkedValue, ok := mapParam["checked"]; ok {
// 					if boolValue, ok := checkedValue.(bool); ok {
// 						processedParam = boolValue
// 					}
// 				}
// 			}

// 		case AmountParam:
// 			// Для AmountParam ожидаем объект {value, dimension}
// 			if mapParam, ok := item.Param.(map[string]interface{}); ok {
// 				amountParam := AmountParam{}

// 				// Обрабатываем поле value
// 				if valueField, ok := mapParam["value"]; ok {
// 					switch vf := valueField.(type) {
// 					case float64:
// 						amountParam.Value = vf
// 					case string:
// 						// Пробуем преобразовать строку в float64
// 						if floatVal, err := strconv.ParseFloat(vf, 64); err == nil {
// 							amountParam.Value = floatVal
// 						}
// 					}
// 				}

// 				// Обрабатываем поле dimension
// 				if dimensionField, ok := mapParam["dimension"]; ok {
// 					if dimension, ok := dimensionField.(string); ok {
// 						amountParam.Dimension = Dimension(dimension)
// 					}
// 				}

// 				processedParam = amountParam
// 			}
// 		}

// 		// Добавляем обработанный параметр в карту
// 		if processedParam != nil {
// 			(*c)[item.Role] = processedParam
// 		} else {
// 			// Если не удалось обработать параметр, используем его как есть
// 			(*c)[item.Role] = item.Param
// 		}
// 	}

// 	return nil
// }
