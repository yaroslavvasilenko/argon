package modules

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
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
			Location: &models.Location{
				ID:   uuid.New().String(),
				Name: "Москва, Россия",
				Area: models.Area{
					Coordinates: models.Coordinates{
						Lat: 55.7558,
						Lng: 37.6173,
					},
					Radius: 10000,
				},
			},
			Categories: []string{"electronics"},
			Characteristics: models.CharacteristicValue{
				models.CHAR_BRAND: models.DropdownOption{
					Value: "Brand1",
					Label: "Brand1",
				},
				models.CHAR_CONDITION: models.DropdownOption{
					Value: "new",
					Label: "Новый",
				},
				models.CHAR_STOCKED: models.CheckboxValue{
					CheckboxValue: true,
				},
				models.CHAR_WEIGHT: models.Amount{
					Value:     2.5,
					Dimension: models.Dimension("kg"),
				},
			},
		}
		resp1 := user.createListing(t, electronicsListing1)
		require.Equal(t, http.StatusOK, resp1.StatusCode, "Объявление 1 должно быть успешно создано")

		// Проверяем, что категория была привязана к объявлению
		var listing1 models.Listing
		json.NewDecoder(resp1.Body).Decode(&listing1)
		t.Logf("Создано объявление 1 с ID: %s", listing1.ID)

		// Создаем еще один товар с другими значениями
		electronicsListing2 := listing.CreateListingRequest{
			Title:       "Test Electronics 2",
			Description: "Test description 2",
			Price:       2000,
			Currency:    models.RUB,
			Location: &models.Location{
				ID:   uuid.New().String(),
				Name: "Санкт-Петербург, Россия",
				Area: models.Area{
					Coordinates: models.Coordinates{
						Lat: 59.9343,
						Lng: 30.3351,
					},
					Radius: 10000,
				},
			},
			Categories: []string{"electronics"},
			Characteristics: models.CharacteristicValue{
				models.CHAR_BRAND: models.DropdownOption{
					Value: "Brand2",
					Label: "Brand2",
				},
				models.CHAR_CONDITION: models.DropdownOption{
					Value: "used",
					Label: "Б/у",
				},
				models.CHAR_STOCKED: models.CheckboxValue{
					CheckboxValue: true,
				},
				models.CHAR_WEIGHT: models.Amount{
					Value:     5.0,
					Dimension: models.Dimension("kg"),
				},
			},
		}
		resp2 := user.createListing(t, electronicsListing2)
		require.Equal(t, http.StatusOK, resp2.StatusCode, "Объявление 2 должно быть успешно создано")

		// Проверяем, что категория была привязана к объявлению
		var listing2 models.Listing
		json.NewDecoder(resp2.Body).Decode(&listing2)
		t.Logf("Создано объявление 2 с ID: %s", listing2.ID)

		// Создаем товар для категории смартфонов с цветом
		smartphoneListing := listing.CreateListingRequest{
			Title:       "Test Smartphone",
			Description: "Test smartphone description",
			Price:       3000,
			Currency:    models.RUB,
			Location: &models.Location{
				ID:   uuid.New().String(),
				Name: "Казань, Россия",
				Area: models.Area{
					Coordinates: models.Coordinates{
						Lat: 55.7887,
						Lng: 49.1221,
					},
					Radius: 10000,
				},
			},
			Categories: []string{"electronics", "smartphones"},
			Characteristics: models.CharacteristicValue{
				models.CHAR_BRAND: models.DropdownOption{
							Value: "Brand3",
							Label: "Brand3",
						},
				// Цвет должен быть структурой ColorParam
				models.CHAR_COLOR: models.Color{
					Color: "red",
				},
				models.CHAR_CONDITION: models.DropdownOption{
							Value: "new",
							Label: "Новый",
						},

				models.CHAR_STOCKED: models.CheckboxValue{
					CheckboxValue: true,
				},
			},
		}
		resp3 := user.createListing(t, smartphoneListing)
		require.Equal(t, http.StatusOK, resp3.StatusCode, "Объявление 3 (смартфон) должно быть успешно создано")

		// Проверяем, что категория была привязана к объявлению
		var listing3 models.Listing
		json.NewDecoder(resp3.Body).Decode(&listing3)
		t.Logf("Создано объявление 3 (смартфон) с ID: %s", listing3.ID)

		// Добавляем небольшую задержку, чтобы дать время на обработку данных
		time.Sleep(100 * time.Millisecond)

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
