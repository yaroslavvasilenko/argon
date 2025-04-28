package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

// ValidateCharacteristics проверяет, что характеристики соответствуют категориям
func ValidateCharacteristicsValue(fl validator.FieldLevel) bool {
	// Получаем значение поля характеристик
	characteristicsValue := fl.Field()
	if characteristicsValue.Kind() != reflect.Map {
		return false
	}

	// Получаем значение поля категорий из родительской структуры
	parentValue := fl.Parent()
	categoriesField := parentValue.FieldByName("Categories")
	if !categoriesField.IsValid() || categoriesField.Kind() != reflect.Slice {
		return false
	}

	categories := make([]string, categoriesField.Len())
	for i := 0; i < categoriesField.Len(); i++ {
		categories[i] = categoriesField.Index(i).String()
	}

	// Получаем все допустимые характеристики для выбранных категорий
	allowedCharacteristics := getAllCharacteristicsForCategories(categories)
	allowedCharacteristicsMap := make(map[string]bool)
	for _, characteristic := range allowedCharacteristics {
		allowedCharacteristicsMap[characteristic] = true
	}

	// Проверяем, что все характеристики допустимы для выбранных категорий
	for _, key := range characteristicsValue.MapKeys() {
		characteristicName := key.String()
		if !allowedCharacteristicsMap[characteristicName] {
			return false
		}

		// Получаем значение характеристики
		characteristicValue := characteristicsValue.MapIndex(key)
		if characteristicValue.IsNil() {
			return false
		}

		if !validateCharacteristicValue(characteristicName, characteristicValue) {
			return false
		}
	}

	return true
}

// validateCharacteristicValue проверяет, что значение характеристики соответствует ожидаемому типу
func validateCharacteristicValue(characteristicName string, interfaceValue reflect.Value) bool {
	paramType, ok := models.CharacteristicValueMap[characteristicName]
	if !ok {
		return false
	}
	characteristicValue := interfaceValue.Elem()

	switch paramType.(type) {
	case models.Color:

		// Проверяем, что значение является Color
		colorValue, ok := characteristicValue.Interface().(models.Color)
		if ok {
			return isValidColor(colorValue.Color)
		}

		// Проверяем, что значение является ColorParam
		_, ok = characteristicValue.Interface().(models.ColorParam)
		if ok {
			return true
		}

		// Проверяем, что значение является строкой
		if characteristicValue.Kind() == reflect.String {
			color := characteristicValue.String()
			if color == "" {
				return false
			}
			return isValidColor(color)
		}

		return false

	case models.DropdownOption:
		// Проверяем, что значение является DropdownOption
		dropdownOption, ok := characteristicValue.Interface().(models.DropdownOption)
		if ok {
			// Проверяем, что есть хотя бы одна опция
			if dropdownOption.Value == "" {
				return false
			}

			// Проверяем, что у первой опции есть значение
			if dropdownOption.Label == "" {
				return false
			}
			return true
		}

		// Проверяем, что значение является StringParam
		stringParam, ok := characteristicValue.Interface().(models.StringParam)
		if ok {
			// Проверяем, что есть хотя бы одна опция
			if len(stringParam.Options) == 0 {
				return false
			}
			return true
		}

		return false
	case models.Amount:
		amountParam, ok := characteristicValue.Interface().(models.Amount)
		if !ok {
			return false
		}

		if !isValidDimension(characteristicName, []models.Dimension{amountParam.Dimension}) {
			return false
		}

		return true

	case models.CheckboxValue:
		// Проверяем, что значение является CheckboxValue
		_, ok := characteristicValue.Interface().(models.CheckboxValue)
		if ok {
			return true
		}

		// Проверяем, что значение является bool
		_, ok = characteristicValue.Interface().(bool)
		if ok {
			return true
		}

		return false
	default:
		return false
	}
}

// isValidColor проверяет, что цвет входит в список допустимых
func isValidColor(color string) bool {
	// Получаем список допустимых цветов из конфига
	cfg := config.GetConfig()

	validColors := cfg.Categories.CategoryOptions[models.CHAR_COLOR]

	// Проверяем наличие цвета в списке
	for _, validColor := range validColors {
		if validColor == color {
			return true
		}
	}

	return false
}

// isValidDimension проверяет, что единица измерения допустима для данной характеристики
func isValidDimension(characteristicName string, dimensions []models.Dimension) bool {
	// Определяем допустимые единицы измерения в зависимости от типа характеристики
	var validDimensions []string

	switch characteristicName {
	case models.CHAR_HEIGHT, models.CHAR_WIDTH, models.CHAR_DEPTH:
		// Для линейных размеров
		validDimensions = []string{models.CM, models.M, models.KM}

	case models.CHAR_AREA:
		// Для площади
		validDimensions = []string{models.CM2, models.M2, models.KM2}

	case models.CHAR_VOLUME:
		// Для объема
		validDimensions = []string{models.CM3, models.M3, models.KM3, models.ML, models.L}

	case models.CHAR_WEIGHT:
		// Для веса
		validDimensions = []string{models.G, models.KG, models.T}

	}

	// Проверяем наличие единицы измерения в списке допустимых
	for _, validDimension := range validDimensions {
		for _, dimension := range dimensions {
			if validDimension == string(dimension) {
				return true
			}
		}
	}

	return false
}

// Получение всех характеристик для категории, включая характеристики родительских категорий
func getAllCharacteristicsForCategories(categories []string) []string {
	allCharacteristics := make(map[string]bool)

	for _, category := range categories {
		if characteristics, ok := config.GetConfig().Categories.CategoryCharacteristics[category]; ok {
			for _, characteristic := range characteristics {
				allCharacteristics[characteristic] = true
			}
		}
	}

	// Преобразуем map в slice
	result := make([]string, 0, len(allCharacteristics))
	for characteristic := range allCharacteristics {
		result = append(result, characteristic)
	}

	return result
}
