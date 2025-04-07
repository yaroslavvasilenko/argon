package listing

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/config"
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
	lang := models.Localization(ctx.Value(models.KeyLanguage).(string))
	if lang == "" {
		lang = models.LanguageDefault
	}

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

type CharacteristicParam []CharacteristicParamItem

type GetCharacteristicsForCategoryResponse struct {
	CharacteristicParams CharacteristicParam `json:"characteristic_params"`
}

// MarshalJSON реализует интерфейс json.Marshaler для типа CharacteristicParam
// func (c CharacteristicParam) MarshalJSON() ([]byte, error) {
// 	if c == nil {
// 		return []byte("null"), nil
// 	}

// 	charItems := make([]CharacteristicParamItem, 0, len(c))
// 	for role, param := range c {
// 		// Определяем тип параметра на основе роли характеристики
// 		paramType, exists := models.CharacteristicParamMap[role]
// 		if !exists {
// 			// Если тип не определен, используем параметр как есть
// 			charItems = append(charItems, CharacteristicParamItem{
// 				Role:  role,
// 				Param: param,
// 			})
// 			continue
// 		}

// 		// Преобразуем параметр в соответствующий тип
// 		var typedParam interface{}
// 		switch paramType {
// 		case models.COLOR_TYPE:
// 			typedParam = &models.ColorParam{}
// 		case models.DROPDOWN_TYPE:
// 			// Для выпадающего списка нужно преобразовать options
// 			if options, ok := param.(map[string]interface{}); ok {
// 				if optionsArray, ok := options["options"].([]interface{}); ok {
// 					stringOptions := make([]string, 0, len(optionsArray))
// 					for _, opt := range optionsArray {
// 						if strOpt, ok := opt.(string); ok {
// 							stringOptions = append(stringOptions, strOpt)
// 						}
// 					}
// 					typedParam = &models.DropdownParam{Options: stringOptions}
// 				} else {
// 					typedParam = &models.DropdownParam{}
// 				}
// 			} else {
// 				typedParam = &models.DropdownParam{}
// 			}
// 		case models.CHECKBOX_TYPE:
// 			typedParam = &models.CheckboxParam{}
// 		case models.DIMENSION_TYPE:
// 			// Для размерных параметров нужно преобразовать dimension_options
// 			if dimensions, ok := param.(map[string]interface{}); ok {
// 				if dimOptions, ok := dimensions["dimension_options"].([]interface{}); ok {
// 					dimensionOptions := make([]models.Dimension, 0, len(dimOptions))
// 					for _, dim := range dimOptions {
// 						if strDim, ok := dim.(string); ok {
// 							dimensionOptions = append(dimensionOptions, models.Dimension(strDim))
// 						}
// 					}
// 					typedParam = &models.DimensionParam{DimensionOptions: dimensionOptions}
// 				} else {
// 					typedParam = &models.DimensionParam{}
// 				}
// 			} else {
// 				typedParam = &models.DimensionParam{}
// 			}
// 		default:
// 			// Если тип не распознан, используем параметр как есть
// 			typedParam = param
// 		}

// 		charItems = append(charItems, CharacteristicParamItem{
// 			Role:  role,
// 			Param: typedParam,
// 		})
// 	}

// 	return json.Marshal(charItems)
// }
