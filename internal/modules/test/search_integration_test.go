package modules

import (
	"fmt"
	"maps"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

func TestSearchListings(t *testing.T) {
	app := createTestApp(t)
	app.cleanDb(t)
	user := app.createUser(t)

	iphone1 := listing.CreateListingRequest{
		Title:       "iPhone 14 Pro",
		Description: "Новый iPhone 14 Pro, 256GB, цвет: космический черный",
		Price:       100001,
		Currency:    models.Currency("RUB"),
	}

	resp := user.createListing(t, iphone1)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	t.Run("Successful search 1 listing", func(t *testing.T) {
		// Выполняем поиск объявления
		req := getSearchListingsRequest("iPhone", 10, "", "relevance", "")

		resp := user.searchListings(t, req)

		require.Len(t, resp.Results, 1)
		foundListing := resp.Results[0]
		assert.Equal(t, iphone1.Title, foundListing.Title)
		assert.Equal(t, iphone1.Description, foundListing.Description)
		assert.Equal(t, iphone1.Price, foundListing.Price)
		assert.Equal(t, iphone1.Currency, foundListing.Currency)
	})

	iphone2 := listing.CreateListingRequest{
		Title:       "iPhone 15 Pro",
		Description: "Новый iPhone 15 Pro, 256GB, цвет: космический черный",
		Price:       100002,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone2)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone3 := listing.CreateListingRequest{
		Title:       "iPhone 16 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100003,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone3)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone4 := listing.CreateListingRequest{
		Title:       "iPhone 17 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100004,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone4)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone5 := listing.CreateListingRequest{
		Title:       "iPhone 18 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100005,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone5)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone6 := listing.CreateListingRequest{
		Title:       "iPhone 19 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100006,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone6)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone7 := listing.CreateListingRequest{
		Title:       "iPhone 20 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100007,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone7)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone8 := listing.CreateListingRequest{
		Title:       "iPhone 21 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100008,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone8)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone9 := listing.CreateListingRequest{
		Title:       "iPhone 22 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100009,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone9)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	t.Run("Successful search 3 listings and check cursor", func(t *testing.T) {
		// Выполняем поиск объявления
		req := getSearchListingsRequest("iPhone", 3, "", "price_asc", "")

		resp := user.searchListings(t, req)

		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone1.Title, resp.Results[0].Title)
		assert.Equal(t, iphone2.Title, resp.Results[1].Title)
		assert.Equal(t, iphone3.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", 3, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", -3, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", -3, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", 3, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone7.Title, resp.Results[0].Title)
		assert.Equal(t, iphone8.Title, resp.Results[1].Title)
		assert.Equal(t, iphone9.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", 3, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		assert.Empty(t, resp.Results)
		assert.Empty(t, resp.CursorAfter)

		req = getSearchListingsRequest("iPhone", -3, "", "price_asc", "")

		resp = user.searchListings(t, req)

		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone7.Title, resp.Results[0].Title)
		assert.Equal(t, iphone8.Title, resp.Results[1].Title)
		assert.Equal(t, iphone9.Title, resp.Results[2].Title)

	})

	t.Run("Successful search by relevance", func(t *testing.T) {
		// Выполняем поиск объявления с сортировкой по релевантности
		req := getSearchListingsRequest("iPhone", 3, "", "relevance", "")

		resp := user.searchListings(t, req)

		// Проверяем, что получили 3 результата
		require.Len(t, resp.Results, 3)

		// Продолжаем пагинацию дальше
		firstPageLastTitle := resp.Results[2].Title
		req = getSearchListingsRequest("iPhone", 3, resp.CursorAfter, "relevance", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		secondPageFirstTitle := resp.Results[0].Title

		// Проверяем, что нет дубликатов между страницами
		assert.NotEqual(t, firstPageLastTitle, secondPageFirstTitle, "Дубликаты в результатах поиска по релевантности")

		// Пагинация в обратном направлении
		req = getSearchListingsRequest("iPhone", -3, resp.CursorAfter, "relevance", resp.SearchID)
		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)

		// Проверяем, что получили предыдущие результаты
		assert.Equal(t, secondPageFirstTitle, resp.Results[0].Title)
	})

	t.Run("Fuzzy search with typo", func(t *testing.T) {
		// Создаем объявление для проверки нечеткого поиска
		tvSamsung := listing.CreateListingRequest{
			Title:       "Samsung Neo QLED TV",
			Description: "Телевизор Samsung Neo QLED, 65 дюймов",
			Price:       120000,
			Currency:    models.Currency("RUB"),
		}
		res := user.createListing(t, tvSamsung)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// Поиск с опечаткой - пропущена буква 'u'
		req := getSearchListingsRequest("Samsng", 5, "", "relevance", "")
		resp := user.searchListings(t, req)

		// Проверяем, что нашли iPhone несмотря на опечатку
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с опечаткой")

		// Проверяем, что все найденные объявления содержат "iPhone"
		for _, listing := range resp.Results {
			assert.Contains(t, listing.Title, "Samsung", "Найдено нерелевантное объявление при поиске с опечаткой")
		}

		// Поиск с другой опечаткой - замена буквы
		req = getSearchListingsRequest("Sumsung", 5, "", "relevance", "")
		resp = user.searchListings(t, req)

		// Проверяем, что нашли Samsung несмотря на опечатку
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с опечаткой Sumsung")

		// Проверяем, что первый результат - Samsung
		assert.Contains(t, resp.Results[0].Title, "Samsung", "На первом месте должен быть Samsung")
	})

	t.Run("Filters search", func(t *testing.T) {
		// Создаем объявление для проверки поиска по характеристикам
		notebook := listing.CreateListingRequest{
			Title:       "ноутбук с характеристиками",
			Description: "ноутбук с характеристиками",
			Price:       120000,
			Currency:    models.Currency("RUB"),
			// Характеристики объявления
			Characteristics: map[string]interface{}{
				models.CHAR_COLOR:   []string{"black", "silver"},
				models.CHAR_BRAND:           "Samsung",
				models.CHAR_STOCKED:            true,
				models.CHAR_WEIGHT:       15,
			},
		}
		res := user.createListing(t, notebook)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// Поиск с фильтрами
		req := getSearchListingsRequest("ноутбук", 5, "", "relevance", "")
		// Добавляем фильтры в запрос поиска
		// Фильтры для поиска
		filters := models.Filters{
			models.PRICE_TYPE: models.PriceFilter{
				Min: 90000,
				Max: 150000,
			},
			models.COLOR_TYPE:    models.ColorFilter{"black", "silver"},
			models.CHAR_BRAND: models.DropdownFilter{"Samsung"},
			models.CHAR_STOCKED: func() models.CheckboxFilter { trueValue := true; return &trueValue }(),
			models.CHAR_WEIGHT: models.DimensionFilter{
				Min:       14,
				Max:       16,
				Dimension: "",
			},
		}
		req.Filters = filters

		resp := user.searchListings(t, req)

		// Проверяем, что нашли ноутбук с заданными фильтрами
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с фильтрами")
		assert.Equal(t, "ноутбук с характеристиками", resp.Results[0].Title, "Найдено неверное объявление")

		// Тест фильтра по цене
		filtersEdit := make(models.Filters)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.PRICE_TYPE] = models.PriceFilter{
			Min: 130000,
			Max: 150000,
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящей ценой
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящей ценой")

		// Тест фильтра по цвету
		filtersEdit = make(models.Filters)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.COLOR_TYPE] = models.ColorFilter{"red", "green"}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим цветом
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим цветом")

		// Тест фильтра по бренду
		filtersEdit = make(models.Filters)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.CHAR_BRAND] = models.DropdownFilter{"Apple", "Lenovo"}
		// Отладочный вывод фильтров
		fmt.Printf("\n\nDEBUG TEST FILTERS: %v\n\n", filtersEdit)
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим брендом
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим брендом")

		// Тест фильтра по наличию на складе
		filtersEdit = make(models.Filters)
		maps.Copy(filtersEdit, filters)
		falseValue := false
		filtersEdit[models.CHAR_STOCKED] = &falseValue
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим состоянием
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим состоянием")

		// Тест фильтра по размеру экрана
		filtersEdit = make(models.Filters)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.CHAR_HEIGHT] = models.DimensionFilter{
			Min:       17,
			Max:       19,
			Dimension: "", // Дюймы по умолчанию
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим размером экрана
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим размером экрана")
	})
}

func TestSearchListingsByLocation(t *testing.T) {
	app := createTestApp(t)
	app.cleanDb(t)
	user := app.createUser(t)

	// Создаем объявление с определенной локацией (Москва, Красная площадь)
	listingWithLocation := listing.CreateListingRequest{
		Title:       "Квартира в центре Москвы",
		Description: "Уютная квартира рядом с Красной площадью",
		Price:       150000,
		Currency:    models.Currency("RUB"),
		Location: models.Location{
			ID:   "moscow_center",
			Name: "Красная площадь",
			Area: models.Area{
				Coordinates: struct {
					Lat float64 `json:"lat" validate:"required"`
					Lng float64 `json:"lng" validate:"required"`
				}{
					Lat: 55.753930, // Координаты Красной площади
					Lng: 37.620795,
				},
				Radius: 1000, // Радиус 1 км
			},
		},
	}

	resp := user.createListing(t, listingWithLocation)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	t.Run("Successful search by location", func(t *testing.T) {
		// Создаем запрос на поиск с теми же координатами
		req := getSearchListingsRequest("Квартира", 10, "", "relevance", "")
		
		// Добавляем локацию в запрос поиска
		req.Location = models.Location{
			Area: models.Area{
				Coordinates: struct {
					Lat float64 `json:"lat" validate:"required"`
					Lng float64 `json:"lng" validate:"required"`
				}{
					Lat: 55.753930, // Те же координаты
					Lng: 37.620795,
				},
				Radius: 1500, // Увеличиваем радиус поиска до 1.5 км
			},
		}

		// Выполняем поиск
		resp := user.searchListings(t, req)

		// Проверяем, что объявление найдено
		require.NotEmpty(t, resp.Results, "Объявление не найдено при поиске по локации")
		require.Len(t, resp.Results, 1, "Найдено неверное количество объявлений")
		assert.Equal(t, listingWithLocation.Title, resp.Results[0].Title)
		assert.Equal(t, listingWithLocation.Description, resp.Results[0].Description)
	})

	t.Run("No results when searching with different location", func(t *testing.T) {
		// Создаем запрос на поиск с другими координатами (Санкт-Петербург)
		req := getSearchListingsRequest("Квартира", 10, "", "relevance", "")
		
		// Добавляем локацию в запрос поиска с другими координатами
		req.Location = models.Location{
			Area: models.Area{
				Coordinates: struct {
					Lat float64 `json:"lat" validate:"required"`
					Lng float64 `json:"lng" validate:"required"`
				}{
					Lat: 59.939095, // Координаты Санкт-Петербурга (Дворцовая площадь)
					Lng: 30.315868,
				},
				Radius: 1500, // Тот же радиус поиска
			},
		}

		// Выполняем поиск
		resp := user.searchListings(t, req)

		// Проверяем, что объявление НЕ найдено
		require.Empty(t, resp.Results, "Объявление найдено при поиске по другой локации")
	})

	t.Run("No results when searching with small radius", func(t *testing.T) {
		// Создаем запрос на поиск с теми же координатами, но маленьким радиусом
		req := getSearchListingsRequest("Квартира", 10, "", "relevance", "")
		
		// Добавляем локацию в запрос поиска с немного смещенными координатами и маленьким радиусом
		req.Location = models.Location{
			Area: models.Area{
				Coordinates: struct {
					Lat float64 `json:"lat" validate:"required"`
					Lng float64 `json:"lng" validate:"required"`
				}{
					Lat: 55.753930 + 0.001, // Смещаем координаты примерно на 100 метров
					Lng: 37.620795 + 0.001,
				},
				Radius: 5, // Очень маленький радиус (5 метров)
			},
		}

		// Выполняем поиск
		resp := user.searchListings(t, req)

		// Проверяем, что объявление НЕ найдено из-за маленького радиуса
		require.Empty(t, resp.Results, "Объявление найдено при поиске с маленьким радиусом")
	})
}
