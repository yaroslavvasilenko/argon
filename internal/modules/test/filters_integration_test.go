package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

func TestGetFiltersForCategory(t *testing.T) {
	app := createTestApp(t)
	user := app.createUser(t)

	app.cleanDb(t)

	t.Run("Check empty filters for empty bd", func(t *testing.T) {
		// Тестируем категорию электроники, которая содержит разные типы характеристик
		categoryId := "electronics"

		// Получаем фильтры для категории
		response, err := user.getFiltersForCategory(t, categoryId, "ru")
		require.NoError(t, err)

		// Проверяем, что фильтры пустые, так как в базе данных нет значений
		assert.Empty(t, response.Filters, "Фильтры должны быть пустыми, если в базе данных нет значений")

		// Проверяем отсутствие фильтра цены (PriceFilterParam)
		_, hasPriceFilter := response.Filters.GetPriceFilter(models.CHAR_PRICE)
		assert.False(t, hasPriceFilter, "Фильтр цены не должен присутствовать при пустой базе данных")

		// Проверяем категорию со смартфонами для фильтра цвета (ColorFilterParam)
		colorResponse, err := user.getFiltersForCategory(t, "smartphones", "ru")
		require.NoError(t, err)

		// Проверяем отсутствие фильтра цвета
		_, hasColorFilter := colorResponse.Filters.GetColorFilter(models.CHAR_COLOR)
		assert.False(t, hasColorFilter, "Фильтр цвета не должен присутствовать при пустой базе данных")

		// Проверяем отсутствие фильтра состояния (DropdownFilterParam)
		_, hasConditionFilter := response.Filters.GetDropdownFilter(models.CHAR_CONDITION)
		assert.False(t, hasConditionFilter, "Фильтр состояния не должен присутствовать при пустой базе данных")

		// Проверяем отсутствие фильтра наличия на складе (CheckboxFilterParam)
		_, hasStockedFilter := response.Filters.GetCheckboxFilter(models.CHAR_STOCKED)
		assert.False(t, hasStockedFilter, "Фильтр наличия на складе не должен присутствовать при пустой базе данных")

		// Проверяем отсутствие фильтра веса (DimensionFilterParam)
		_, hasWeightFilter := response.Filters.GetDimensionFilter(models.CHAR_WEIGHT)
		assert.False(t, hasWeightFilter, "Фильтр веса не должен присутствовать при пустой базе данных")

		// Проверяем, что в ответе нет никаких фильтров
		assert.Zero(t, len(response.Filters), "В ответе не должно быть никаких фильтров при пустой базе данных")
	})

	t.Run("Получение фильтров для категории электроники", func(t *testing.T) {
		// Создаем тестовые товары для категории электроники
		// Создаем товар с ценой, брендом, состоянием, наличием на складе и весом
		electronicsListing1 := listing.CreateListingRequest{
			Title:       "Test Electronics 1",
			Description: "Test description 1",
			Price:       1000,
			Currency:    models.RUB,
			Categories:  []string{"electronics"},
			Characteristics: map[string]interface{}{
				models.CHAR_BRAND:     []string{"Brand1"},
				models.CHAR_CONDITION: []string{"new"},
				models.CHAR_STOCKED:   true,
				models.CHAR_WEIGHT:    2.5,
			},
		}
		user.createListing(t, electronicsListing1)

		// Создаем еще один товар с другими значениями
		electronicsListing2 := listing.CreateListingRequest{
			Title:       "Test Electronics 2",
			Description: "Test description 2",
			Price:       2000,
			Currency:    models.RUB,
			Categories:  []string{"electronics"},
			Characteristics: map[string]interface{}{
				models.CHAR_BRAND:     []string{"Brand2"},
				models.CHAR_CONDITION: []string{"used"},
				models.CHAR_STOCKED:   false,
				models.CHAR_WEIGHT:    1.5,
			},
		}
		user.createListing(t, electronicsListing2)

		// Создаем товар для категории смартфонов с цветом
		smartphoneListing := listing.CreateListingRequest{
			Title:       "Test Smartphone",
			Description: "Test smartphone description",
			Price:       3000,
			Currency:    models.RUB,
			Categories:  []string{"smartphones"},
			Characteristics: map[string]interface{}{
				models.CHAR_BRAND:     []string{"Brand3"},
				models.CHAR_COLOR:     []string{"red"},
				models.CHAR_CONDITION: []string{"new"},
				models.CHAR_STOCKED:   true,
			},
		}
		user.createListing(t, smartphoneListing)

		// Вызываем API для получения фильтров
		response, err := user.getFiltersForCategory(t, "electronics", "ru")
		require.NoError(t, err)
		filters := response.Filters

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
		response, err := user.getFiltersForCategory(t, "nonexistent_category", "ru")
		require.NoError(t, err)

		// Проверяем, что фильтры пустые
		assert.Empty(t, response.Filters, "Фильтры должны быть пустыми для несуществующей категории")
	})
}
