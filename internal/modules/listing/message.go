package listing

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/core/parser"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// SearchListingsResponse represents a response to a search listings request.

type ResponseGetCategories struct {
	Categories []CategoryNode `json:"categories"`
}

type CategoryNode struct {
	Category      Category       `json:"category"`
	Subcategories []CategoryNode `json:"subcategories,omitempty"`
}

type Category struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Image *string `json:"image,omitempty"`
}

type CreateListingRequest struct {
	Title           string                `json:"title" validate:"required"`
	Description     string                `json:"description,omitempty"`
	Price           float64               `json:"price,omitempty" validate:"gte=0"`
	Currency        models.Currency       `json:"currency,omitempty" validate:"required,oneof=USD EUR RUB ARS"`
	Location        *models.Location      `json:"location,omitempty"`
	Categories      []string              `json:"categories,omitempty" validate:"required,categories_validation"`
	Characteristics models.Characteristic `json:"characteristics,omitempty" validate:"characteristics_validation"`
	Images          []string              `json:"images"`
}

type CreateListingResponse struct {
	ID              uuid.UUID              `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description,omitempty"`
	Price           float64                `json:"price,omitempty"`
	Currency        models.Currency        `json:"currency,omitempty"`
	Location        models.Location        `json:"location,omitempty"`
	Categories      []Category             `json:"categories"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty"`
	Boosts          []BoostResp            `json:"boosts,omitempty"`
	Images          []string               `json:"images"`
}

type BoostResp struct {
	Type              models.BoostType `json:"type"`
	CommissionPercent float64          `json:"commission_percent"`
}

type UpdateListingRequest struct {
	ID              uuid.UUID              `json:"id" validate:"required"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description,omitempty"`
	Price           float64                `json:"price,omitempty" validate:"gte=0"`
	Currency        models.Currency        `json:"currency,omitempty" validate:"required,oneof=USD EUR RUB"`
	Location        models.Location        `json:"location,omitempty"`
	Categories      []string               `json:"categories,omitempty" validate:"categories_validation"`
	Characteristics map[string]interface{} `json:"characteristics,omitempty" validate:"characteristics_validation"`
	Boosts          []BoostResp            `json:"boosts,omitempty"`
	Images          []string               `json:"images"`
}

type FullListingResponse struct {
	ID                  uuid.UUID             `json:"id"`
	Title               string                `json:"title"`
	Description         string                `json:"description"`
	OriginalDescription string                `json:"original_description"`
	Price               float64               `json:"price"`
	Currency            models.Currency       `json:"currency"`
	OriginalPrice       float64               `json:"original_price"`
	OriginalCurrency    models.Currency       `json:"original_currency"`
	Location            models.Location       `json:"location"`
	Seller              models.Seller         `json:"seller"`
	Categories          []Category            `json:"categories"`
	Characteristics     models.Characteristic `json:"characteristics"`
	Images              []string              `json:"images"`
	CreatedAt           int64                 `json:"created_at"`
	UpdatedAt           int64                 `json:"updated_at"`
	Boosts              []BoostResp           `json:"boosts,omitempty"`
	IsEditable          bool                  `json:"is_editable"`
	IsBuyable           bool                  `json:"is_buyable"`
	IsNSFW              bool                  `json:"is_nsfw"`
}

type GetFiltersForCategoryResponse struct {
	Filters models.Filters `json:"filter_params"`
}

// getCategoriesWithLocalizedNames получает локализованные названия категорий
func GetCategoriesWithLocalizedNames(ctx context.Context, categoryIDs []string) ([]Category, error) {
	// Получаем язык из контекста
	lang := parser.GetLang(ctx)

	// Загружаем локализации
	var translations map[string]string
	var langData string

	switch lang {
	case models.LanguageRu:
		langData = config.GetConfig().Categories.LangCategories.Ru
	case models.LanguageEn:
		langData = config.GetConfig().Categories.LangCategories.En
	case models.LanguageEs:
		langData = config.GetConfig().Categories.LangCategories.Es
	default:
		return nil, errors.New("не поддерживаемый язык: " + string(lang))
	}

	// Распаковываем локализации
	if err := json.Unmarshal([]byte(langData), &translations); err != nil {
		return nil, errors.New("ошибка при разборе локализаций: " + err.Error())
	}

	// Создаем массив категорий с локализованными названиями
	categories := make([]Category, 0, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		name, ok := translations[categoryID]
		if !ok {
			name = categoryID // Если перевод не найден, используем ID как имя
		}
		categories = append(categories, Category{
			ID:   categoryID,
			Name: name,
		})
	}

	return categories, nil
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler для типа Characteristic
// func (c *CharacteristicParam) UnmarshalJSON(data []byte) error {
// 	var charItems []CharacteristicParamItem
// 	if err := json.Unmarshal(data, &charItems); err != nil {
// 		return err
// 	}

// 	if *c == nil {
// 		*c = make(CharacteristicParam)
// 	}

// 	for _, item := range charItems {
// 		(*c)[item.Role] = item.Param
// 	}

// 	return nil
// }

type CharacteristicParamItem struct {
	Role  string      `json:"role"`
	Param interface{} `json:"param"`
}

type Option struct {
	Options []CharacteristicParamItem `json:"options"`
}

type GetCharacteristicsForCategoryResponse struct {
	Option Option `json:"characteristic_params"`
}

// MarshalJSON реализует интерфейс json.Marshaler для типа Option
func (c Option) MarshalJSON() ([]byte, error) {
	if c.Options == nil {
		return []byte("null"), nil
	}

	charItems := make([]CharacteristicParamItem, 0, len(c.Options))
	for _, item := range c.Options {
		role := item.Role
		param := item.Param

		// Определяем тип параметра на основе роли характеристики
		paramType, exists := models.CharacteristicParamMap[role]
		if !exists {
			// Если тип не определен, используем параметр как есть
			charItems = append(charItems, CharacteristicParamItem{
				Role:  role,
				Param: param,
			})
			continue
		}

		// Преобразуем параметр в соответствующий тип
		var typedParam interface{}
		switch paramType.(type) {
		case models.ColorParam:
			typedParam = &models.ColorParam{}
		case models.StringParam:
			// Для выпадающего списка нужно преобразовать options
			stringParam := models.StringParam{
				Options: make([]models.DropdownOptionItem, 0),
			}

			// Пытаемся получить опции из разных типов параметров
			switch p := param.(type) {
			case map[string]interface{}:
				// Если это карта, пытаемся получить поле options
				if optionsField, ok := p["options"]; ok {
					switch opts := optionsField.(type) {
					case []interface{}:
						// Обрабатываем массив интерфейсов
						for _, opt := range opts {
							switch o := opt.(type) {
							case map[string]interface{}:
								// Если это объект, извлекаем value и label
								option := models.DropdownOptionItem{}
								if value, ok := o["value"].(string); ok {
									option.Value = value
								}
								if label, ok := o["label"].(string); ok {
									option.Label = label
								}
								stringParam.Options = append(stringParam.Options, option)
							case string:
								// Если это строка, используем её как value
								stringParam.Options = append(stringParam.Options, models.DropdownOptionItem{Value: o})
							}
						}
					case []string:
						// Если это массив строк, преобразуем каждую в опцию
						for _, value := range opts {
							stringParam.Options = append(stringParam.Options, models.DropdownOptionItem{Value: value})
						}
					}
				}
			case models.StringParam:
				// Если это уже StringParam, используем его напрямую
				stringParam = p
			case []models.DropdownOptionItem:
				// Если это массив опций, используем его как Options
				stringParam.Options = p
			case []interface{}:
				// Если это массив интерфейсов, преобразуем каждый элемент
				for _, item := range p {
					switch i := item.(type) {
					case map[string]interface{}:
						option := models.DropdownOptionItem{}
						if value, ok := i["value"].(string); ok {
							option.Value = value
						}
						if label, ok := i["label"].(string); ok {
							option.Label = label
						}
						stringParam.Options = append(stringParam.Options, option)
					case string:
						stringParam.Options = append(stringParam.Options, models.DropdownOptionItem{Value: i})
					}
				}
			}

			typedParam = stringParam
		case models.CheckboxParam:
			typedParam = &models.CheckboxParam{}
		case models.AmountParam:
			// Для размерных параметров нужно преобразовать dimension_options
			// Инициализируем AmountParam с пустым массивом DimensionOptions
			amountParam := models.AmountParam{
				DimensionOptions: make([]models.Dimension, 0),
			}

			// Пытаемся получить значение и dimension_options из параметра
			if mapParam, ok := param.(map[string]interface{}); ok {
				// Обрабатываем поле value
				if valueField, ok := mapParam["value"]; ok {
					switch vf := valueField.(type) {
					case float64:
						amountParam.Value = vf
					}
				}

				// Обрабатываем поле dimension_options
				if dimensionField, ok := mapParam["dimension_options"]; ok {
					switch df := dimensionField.(type) {
					case []interface{}:
						for _, dim := range df {
							if strDim, ok := dim.(string); ok {
								amountParam.DimensionOptions = append(amountParam.DimensionOptions, models.Dimension(strDim))
							}
						}
					case []string:
						for _, dim := range df {
							amountParam.DimensionOptions = append(amountParam.DimensionOptions, models.Dimension(dim))
						}
					case string:
						// Если пришла одиночная строка, добавляем её как единственный элемент массива
						amountParam.DimensionOptions = append(amountParam.DimensionOptions, models.Dimension(df))
					}
				}
			} else if amountP, ok := param.(models.AmountParam); ok {
				// Если параметр уже имеет нужный тип, просто используем его
				amountParam = amountP
			}

			typedParam = amountParam
		default:
			// Если тип не распознан, используем параметр как есть
			typedParam = param
		}

		charItems = append(charItems, CharacteristicParamItem{
			Role:  role,
			Param: typedParam,
		})
	}

	return json.Marshal(charItems)
}
