package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFiltersUnmarshalJSON(t *testing.T) {
	// Тест для проверки корректного разбора JSON с различными типами фильтров
	t.Run("Разбор JSON с различными типами фильтров", func(t *testing.T) {
		// JSON с разными типами фильтров
		jsonData := `[
			{"role": "price", "param": {"min": 100, "max": 1000}},
			{"role": "color", "param": "white"},
			{"role": "brand", "param": "samsung"},
			{"role": "condition", "param": ["new", "used"]},
			{"role": "stocked", "param": true},
			{"role": "height", "param": {"min": 10, "max": 50, "dimension": "cm"}}
		]`

		var filters Filters
		err := json.Unmarshal([]byte(jsonData), &filters)
		require.NoError(t, err, "Ошибка при разборе JSON")

		// Проверяем фильтр цены
		priceFilter, ok := filters.GetPriceFilter(CHAR_PRICE)
		require.True(t, ok, "Фильтр цены не найден")
		assert.Equal(t, 100, priceFilter.Min, "Минимальная цена должна быть 100")
		assert.Equal(t, 1000, priceFilter.Max, "Максимальная цена должна быть 1000")

		// Проверяем фильтр цвета (строковое значение)
		colorFilter, ok := filters.GetColorFilter(CHAR_COLOR)
		require.True(t, ok, "Фильтр цвета не найден")
		require.Len(t, colorFilter, 1, "Фильтр цвета должен содержать 1 элемент")
		assert.Equal(t, "white", colorFilter[0], "Цвет должен быть 'white'")

		// Проверяем фильтр бренда (строковое значение)
		brandFilter, ok := filters.GetDropdownFilter(CHAR_BRAND)
		require.True(t, ok, "Фильтр бренда не найден")
		require.Len(t, brandFilter, 1, "Фильтр бренда должен содержать 1 элемент")
		assert.Equal(t, "samsung", brandFilter[0], "Бренд должен быть 'samsung'")

		// Проверяем фильтр состояния (массив строк)
		conditionFilter, ok := filters.GetDropdownFilter(CHAR_CONDITION)
		require.True(t, ok, "Фильтр состояния не найден")
		require.Len(t, conditionFilter, 2, "Фильтр состояния должен содержать 2 элемента")
		assert.Equal(t, "new", conditionFilter[0], "Первое состояние должно быть 'new'")
		assert.Equal(t, "used", conditionFilter[1], "Второе состояние должно быть 'used'")

		// Проверяем фильтр наличия (булево значение)
		stockedFilter, ok := filters.GetCheckboxFilter(CHAR_STOCKED)
		require.True(t, ok, "Фильтр наличия не найден")
		require.NotNil(t, stockedFilter, "Фильтр наличия не должен быть nil")
		assert.True(t, *stockedFilter, "Значение наличия должно быть true")

		// Проверяем фильтр высоты (размерное значение)
		heightFilter, ok := filters.GetDimensionFilter(CHAR_HEIGHT)
		require.True(t, ok, "Фильтр высоты не найден")
		assert.Equal(t, 10, heightFilter.Min, "Минимальная высота должна быть 10")
		assert.Equal(t, 50, heightFilter.Max, "Максимальная высота должна быть 50")
		assert.Equal(t, "cm", heightFilter.Dimension, "Единица измерения должна быть 'cm'")
	})

	// Тест для проверки сериализации и десериализации
	t.Run("Сериализация и десериализация фильтров", func(t *testing.T) {
		// Создаем фильтры
		filters := make(Filters)

		// Добавляем фильтр цены
		priceFilter := PriceFilter{Min: 200, Max: 2000}
		filters[CHAR_PRICE] = priceFilter

		// Добавляем фильтр цвета
		colorFilter := ColorFilter{"black", "red"}
		filters[CHAR_COLOR] = colorFilter

		// Добавляем фильтр бренда
		brandFilter := DropdownFilter{"apple", "xiaomi"}
		filters[CHAR_BRAND] = brandFilter

		// Добавляем фильтр наличия
		stockedValue := true
		checkboxFilter := CheckboxFilter(&stockedValue)
		filters[CHAR_STOCKED] = checkboxFilter

		// Добавляем фильтр размера
		dimensionFilter := DimensionFilter{Min: 5, Max: 30, Dimension: "cm"}
		filters[CHAR_HEIGHT] = dimensionFilter

		// Сериализуем фильтры в JSON
		jsonData, err := json.Marshal(filters)
		require.NoError(t, err, "Ошибка при сериализации фильтров")

		// Десериализуем JSON обратно в фильтры
		var newFilters Filters
		err = json.Unmarshal(jsonData, &newFilters)
		require.NoError(t, err, "Ошибка при десериализации фильтров")

		// Проверяем, что фильтры совпадают
		assert.Len(t, newFilters, len(filters), "Количество фильтров должно совпадать")

		// Проверяем фильтр цены
		newPriceFilter, ok := newFilters.GetPriceFilter(CHAR_PRICE)
		require.True(t, ok, "Фильтр цены не найден после десериализации")
		assert.Equal(t, priceFilter.Min, newPriceFilter.Min, "Минимальная цена должна совпадать")
		assert.Equal(t, priceFilter.Max, newPriceFilter.Max, "Максимальная цена должна совпадать")

		// Проверяем фильтр цвета
		newColorFilter, ok := newFilters.GetColorFilter(CHAR_COLOR)
		require.True(t, ok, "Фильтр цвета не найден после десериализации")
		assert.ElementsMatch(t, colorFilter, newColorFilter, "Значения цвета должны совпадать")

		// Проверяем фильтр бренда
		newBrandFilter, ok := newFilters.GetDropdownFilter(CHAR_BRAND)
		require.True(t, ok, "Фильтр бренда не найден после десериализации")
		assert.ElementsMatch(t, brandFilter, newBrandFilter, "Значения бренда должны совпадать")

		// Проверяем фильтр наличия
		newStockedFilter, ok := newFilters.GetCheckboxFilter(CHAR_STOCKED)
		require.True(t, ok, "Фильтр наличия не найден после десериализации")
		require.NotNil(t, newStockedFilter, "Фильтр наличия не должен быть nil")
		assert.Equal(t, true, *newStockedFilter, "Значение наличия должно совпадать")

		// Проверяем фильтр размера
		newDimensionFilter, ok := newFilters.GetDimensionFilter(CHAR_HEIGHT)
		require.True(t, ok, "Фильтр размера не найден после десериализации")
		assert.Equal(t, dimensionFilter.Min, newDimensionFilter.Min, "Минимальный размер должен совпадать")
		assert.Equal(t, dimensionFilter.Max, newDimensionFilter.Max, "Максимальный размер должен совпадать")
		assert.Equal(t, dimensionFilter.Dimension, newDimensionFilter.Dimension, "Единица измерения должна совпадать")
	})

	// Тест для проверки обработки JSON в формате характеристик товара
	t.Run("Разбор JSON в формате характеристик товара", func(t *testing.T) {
		// JSON в формате характеристик товара
		jsonData := `[
			{"role": "price", "value": 500},
			{"role": "color", "value": "white"},
			{"role": "brand", "value": "samsung"},
			{"role": "condition", "value": ["new", "used"]},
			{"role": "stocked", "value": true},
			{"role": "height", "value": {"min": 10, "max": 50, "dimension": "cm"}}
		]`

		// Преобразуем JSON в формат, который ожидает наш UnmarshalJSON
		var characteristics []struct {
			Role  string          `json:"role"`
			Value json.RawMessage `json:"value"`
		}
		err := json.Unmarshal([]byte(jsonData), &characteristics)
		require.NoError(t, err, "Ошибка при разборе JSON характеристик")

		// Создаем JSON в формате, который ожидает наш UnmarshalJSON
		filtersData := make([]struct {
			Role  string          `json:"role"`
			Param json.RawMessage `json:"param"`
		}, len(characteristics))

		for i, c := range characteristics {
			filtersData[i] = struct {
				Role  string          `json:"role"`
				Param json.RawMessage `json:"param"`
			}{
				Role:  c.Role,
				Param: c.Value,
			}
		}

		// Сериализуем в JSON
		filtersJSON, err := json.Marshal(filtersData)
		require.NoError(t, err, "Ошибка при сериализации фильтров")

		// Десериализуем JSON в фильтры
		var filters Filters
		err = json.Unmarshal(filtersJSON, &filters)
		require.NoError(t, err, "Ошибка при десериализации фильтров")

		// Проверяем фильтр цвета
		colorFilter, ok := filters.GetColorFilter(CHAR_COLOR)
		require.True(t, ok, "Фильтр цвета не найден")
		require.Len(t, colorFilter, 1, "Фильтр цвета должен содержать 1 элемент")
		assert.Equal(t, "white", colorFilter[0], "Цвет должен быть 'white'")

		// Проверяем фильтр бренда
		brandFilter, ok := filters.GetDropdownFilter(CHAR_BRAND)
		require.True(t, ok, "Фильтр бренда не найден")
		require.Len(t, brandFilter, 1, "Фильтр бренда должен содержать 1 элемент")
		assert.Equal(t, "samsung", brandFilter[0], "Бренд должен быть 'samsung'")
	})
}
