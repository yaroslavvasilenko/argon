package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

func TestGetFiltersForCategory(t *testing.T) {
	app := createTestApp(t)
	user := app.createUser(t)

	t.Run("Получение фильтров для категории электроники", func(t *testing.T) {
		// Вызываем API для получения фильтров
		filters, err := user.getFiltersForCategory(t, "electronics", "ru")
		require.NoError(t, err)

		// Проверяем, что фильтры не пустые
		require.NotEmpty(t, filters, "Фильтры не должны быть пустыми")

		// Проверяем наличие ожидаемых фильтров для категории electronics
		expectedFilters := []string{"price", "brand", "condition", "stocked", "weight"}
		for _, expectedFilter := range expectedFilters {
			_, exists := filters[expectedFilter]
			assert.True(t, exists, "Фильтр '%s' должен присутствовать в ответе", expectedFilter)
		}

		// Проверяем типы значений фильтров
		priceFilter, hasPriceFilter := filters["price"]
		if assert.True(t, hasPriceFilter, "Фильтр 'price' должен присутствовать") {
			_, ok := priceFilter.(models.PriceFilter)
			assert.True(t, ok, "Значение фильтра 'price' должно быть типа PriceFilter")
		}

		brandFilter, hasBrandFilter := filters["brand"]
		if assert.True(t, hasBrandFilter, "Фильтр 'brand' должен присутствовать") {
			_, ok := brandFilter.(models.DropdownFilter)
			assert.True(t, ok, "Значение фильтра 'brand' должно быть типа DropdownFilter")
		}

		conditionFilter, hasConditionFilter := filters["condition"]
		if assert.True(t, hasConditionFilter, "Фильтр 'condition' должен присутствовать") {
			_, ok := conditionFilter.(models.DropdownFilter)
			assert.True(t, ok, "Значение фильтра 'condition' должно быть типа DropdownFilter")
		}

		stockedFilter, hasStockedFilter := filters["stocked"]
		if assert.True(t, hasStockedFilter, "Фильтр 'stocked' должен присутствовать") {
			_, ok := stockedFilter.(models.CheckboxFilter)
			assert.True(t, ok, "Значение фильтра 'stocked' должно быть типа CheckboxFilter")
		}

		weightFilter, hasWeightFilter := filters["weight"]
		if assert.True(t, hasWeightFilter, "Фильтр 'weight' должен присутствовать") {
			_, ok := weightFilter.(models.DimensionFilter)
			assert.True(t, ok, "Значение фильтра 'weight' должно быть типа DimensionFilter")
		}
	})

	t.Run("Получение фильтров для несуществующей категории", func(t *testing.T) {
		// Вызываем API для получения фильтров для несуществующей категории
		filters, err := user.getFiltersForCategory(t, "nonexistent_category", "ru")
		require.NoError(t, err)

		// Проверяем, что фильтры пустые
		assert.Empty(t, filters, "Фильтры должны быть пустыми для несуществующей категории")
	})
}
