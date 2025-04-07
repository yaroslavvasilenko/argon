package modules

import (
	"encoding/json"
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

		req = getSearchListingsRequest("iPhone", 3, *resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", -3, *resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", -3, *resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone4.Title, resp.Results[0].Title)
		assert.Equal(t, iphone5.Title, resp.Results[1].Title)
		assert.Equal(t, iphone6.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", 3, *resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		assert.Equal(t, iphone7.Title, resp.Results[0].Title)
		assert.Equal(t, iphone8.Title, resp.Results[1].Title)
		assert.Equal(t, iphone9.Title, resp.Results[2].Title)

		req = getSearchListingsRequest("iPhone", 3, *resp.CursorAfter, "price_asc", resp.SearchID)

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
		req = getSearchListingsRequest("iPhone", 3, *resp.CursorAfter, "relevance", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 3)
		secondPageFirstTitle := resp.Results[0].Title

		// Проверяем, что нет дубликатов между страницами
		assert.NotEqual(t, firstPageLastTitle, secondPageFirstTitle, "Дубликаты в результатах поиска по релевантности")

		// Пагинация в обратном направлении
		req = getSearchListingsRequest("iPhone", -3, *resp.CursorAfter, "relevance", resp.SearchID)
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
				models.CHAR_BRAND:   "Samsung",
				models.CHAR_STOCKED: true,
				models.CHAR_WEIGHT:  15,
			},
		}
		res := user.createListing(t, notebook)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// Поиск с фильтрами
		req := getSearchListingsRequest("ноутбук", 5, "", "relevance", "")
		// Добавляем фильтры в запрос поиска
		// Фильтры для поиска
		filters := models.FilterParams{
			models.PRICE_TYPE: models.FilterItem{
				Role:  models.PRICE_TYPE,
				Param: models.PriceFilter{
					Min: 90000,
					Max: 150000,
				},
			},
			models.COLOR_TYPE:   models.FilterItem{Role: models.COLOR_TYPE, Param: models.ColorFilter{Options: []string{"black", "silver"}}},
			models.CHAR_BRAND:   models.FilterItem{Role: models.CHAR_BRAND, Param: models.DropdownFilter{"Samsung"}},
			models.CHAR_STOCKED: models.FilterItem{Role: models.CHAR_STOCKED, Param: func() models.CheckboxFilter { trueValue := true; return &trueValue }()},
			models.CHAR_WEIGHT: models.FilterItem{Role: models.CHAR_WEIGHT, Param: models.DimensionFilter{
				Min:       14,
				Max:       16,
				Dimension: "",
			}},
		}
		req.Filters = filters

		resp := user.searchListings(t, req)

		// Проверяем, что нашли ноутбук с заданными фильтрами
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с фильтрами")
		assert.Equal(t, "ноутбук с характеристиками", resp.Results[0].Title, "Найдено неверное объявление")

		// Тест фильтра по цене
		filtersEdit := make(models.FilterParams)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.PRICE_TYPE] = models.FilterItem{
			Role:  models.PRICE_TYPE,
			Param: models.PriceFilter{
				Min: 130000,
				Max: 150000,
			},
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящей ценой
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящей ценой")

		// Тест фильтра по цвету
		filtersEdit = make(models.FilterParams)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.COLOR_TYPE] = models.FilterItem{
			Role:  models.COLOR_TYPE,
			Param: models.ColorFilter{Options: []string{"red", "green"}},
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим цветом
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим цветом")

		// Тест фильтра по бренду
		filtersEdit = make(models.FilterParams)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.CHAR_BRAND] = models.FilterItem{
			Role:  models.CHAR_BRAND,
			Param: models.DropdownFilter{"Apple", "Lenovo"},
		}
		// Отладочный вывод фильтров
		fmt.Printf("\n\nDEBUG TEST FILTERS: %v\n\n", filtersEdit)
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим брендом
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим брендом")

		// Тест фильтра по наличию на складе
		filtersEdit = make(models.FilterParams)
		maps.Copy(filtersEdit, filters)
		falseValue := false
		filtersEdit[models.CHAR_STOCKED] = models.FilterItem{
			Role:  models.CHAR_STOCKED,
			Param: &falseValue,
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим состоянием
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим состоянием")

		// Тест фильтра по размеру экрана
		filtersEdit = make(models.FilterParams)
		maps.Copy(filtersEdit, filters)
		filtersEdit[models.CHAR_HEIGHT] = models.FilterItem{
			Role:  models.CHAR_HEIGHT,
			Param: models.DimensionFilter{
				Min:       17,
				Max:       19,
				Dimension: "", // Дюймы по умолчанию
			},
		}
		req.Filters = filtersEdit
		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено при поиске с неподходящим размером экрана
		require.Empty(t, resp.Results, "Найдено объявление при поиске с неподходящим размером экрана")
	})

	t.Run("Filter with key but no value", func(t *testing.T) {
		// Создаем объявление для проверки поиска по характеристикам
		emptyFilterNotebook := listing.CreateListingRequest{
			Title:       "ноутбук с пустым фильтром",
			Description: "ноутбук для теста пустых фильтров",
			Price:       130000,
			Currency:    models.Currency("RUB"),
			// Характеристики объявления
			Characteristics: map[string]interface{}{
				models.CHAR_COLOR:   []string{"white", "gold"},
				models.CHAR_BRAND:   "Apple",
				models.CHAR_STOCKED: true,
				models.CHAR_WEIGHT:  20,
			},
		}
		res := user.createListing(t, emptyFilterNotebook)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// Поиск с пустым фильтром
		req := getSearchListingsRequest("ноутбук", 5, "", "relevance", "")

		// Создаем фильтр с ключом, но без значения
		// Проверяем каждый тип фильтра

		// Тест пустого фильтра цены
		filters := models.FilterParams{
			models.PRICE_TYPE: models.FilterItem{
				Role:  models.PRICE_TYPE,
				Param: models.PriceFilter{},
			},
		}
		req.Filters = filters
		resp := user.searchListings(t, req)

		// Проверяем, что нашли ноутбук, так как пустой фильтр цены не должен влиять на поиск
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым фильтром цены")

		// Тест пустого фильтра цвета
		filters = models.FilterParams{
			models.COLOR_TYPE: models.FilterItem{
				Role:  models.COLOR_TYPE,
				Param: models.ColorFilter{Options: []string{}},
			},
		}
		req.Filters = filters
		resp = user.searchListings(t, req)

		// Проверяем, что нашли ноутбук, так как пустой фильтр цвета не должен влиять на поиск
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым фильтром цвета")

		// Тест пустого фильтра бренда
		filters = models.FilterParams{
			models.CHAR_BRAND: models.FilterItem{
				Role:  models.CHAR_BRAND,
				Param: models.DropdownFilter{},
			},
		}
		req.Filters = filters
		resp = user.searchListings(t, req)

		// Проверяем, что нашли ноутбук, так как пустой фильтр бренда не должен влиять на поиск
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым фильтром бренда")

		// Тест пустого фильтра размеров
		filters = models.FilterParams{
			models.CHAR_WEIGHT: models.FilterItem{
				Role:  models.CHAR_WEIGHT,
				Param: models.DimensionFilter{
					Min:       0,
					Max:       0,
					Dimension: "kg", // Указываем размерность, так как она обязательна
				},
			},
		}
		req.Filters = filters
		resp = user.searchListings(t, req)

		// Проверяем, что нашли ноутбук, так как пустой фильтр веса не должен влиять на поиск
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым фильтром веса")

		// Тест фильтра с nil значением
		filters = models.FilterParams{
			models.CHAR_STOCKED: models.FilterItem{
				Role:  models.CHAR_STOCKED,
				Param: nil, // Nil значение для булевого фильтра
			},
		}
		req.Filters = filters
		resp = user.searchListings(t, req)

		// Проверяем, что нашли ноутбук, так как nil фильтр не должен влиять на поиск
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с nil фильтром")
	})

	t.Run("Empty query search", func(t *testing.T) {
		// Создаем объявление для проверки поиска с пустым запросом
		emptyQueryItem := listing.CreateListingRequest{
			Title:       "Тестовый товар для поиска с пустым запросом",
			Description: "Этот товар должен находиться при поиске с пустым запросом",
			Price:       50000,
			Currency:    models.Currency("RUB"),
			// Характеристики объявления
			Characteristics: map[string]interface{}{
				models.CHAR_COLOR:   []string{"black"},
				models.CHAR_BRAND:   "TestBrand",
				models.CHAR_STOCKED: true,
				models.CHAR_WEIGHT:  15,
			},
		}
		res := user.createListing(t, emptyQueryItem)
		require.Equal(t, http.StatusOK, res.StatusCode)

		// Поиск с пустым запросом
		req := getSearchListingsRequest("", 10, "", "", "")

		// Выполняем поиск без фильтров
		resp := user.searchListings(t, req)

		// Проверяем, что поиск вернул результаты
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым запросом")

		// Проверяем, что среди результатов есть наш тестовый товар
		found := false
		for _, result := range resp.Results {
			if result.Title == "Тестовый товар для поиска с пустым запросом" {
				found = true
				break
			}
		}
		require.True(t, found, "Тестовый товар не найден при поиске с пустым запросом")

		// Проверяем поиск с пустым запросом и фильтром
		filters := models.FilterParams{
			models.CHAR_BRAND: models.FilterItem{
				Role:  models.CHAR_BRAND,
				Param: models.DropdownFilter{"TestBrand"},
			},
		}
		req.Filters = filters
		resp = user.searchListings(t, req)

		// Проверяем, что поиск вернул результаты
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске с пустым запросом и фильтром бренда")

		// Проверяем, что среди результатов есть наш тестовый товар
		found = false
		for _, result := range resp.Results {
			if result.Title == "Тестовый товар для поиска с пустым запросом" {
				found = true
				break
			}
		}
		require.True(t, found, "Тестовый товар не найден при поиске с пустым запросом и фильтром бренда")
	})

	t.Run("Search by category filter", func(t *testing.T) {
		// Создаем объявления в разных категориях
		smartphone := listing.CreateListingRequest{
			Title:       "iPhone 13 Pro Max",
			Description: "Смартфон Apple iPhone 13 Pro Max",
			Price:       90000,
			Currency:    models.Currency("RUB"),
			Categories:  []string{"smartphones"}, // Категория смартфонов
		}
		res := user.createListing(t, smartphone)
		require.Equal(t, http.StatusOK, res.StatusCode)
		
		// Получаем ID созданного объявления смартфона
		var smartphoneResp listing.CreateListingResponse
		err := json.NewDecoder(res.Body).Decode(&smartphoneResp)
		require.NoError(t, err)
		t.Logf("Создано объявление смартфона с ID: %s, категории: %v", smartphoneResp.ID, smartphoneResp.Categories)

		clothing := listing.CreateListingRequest{
			Title:       "Мужская куртка",
			Description: "Стильная мужская куртка",
			Price:       5000,
			Currency:    models.Currency("RUB"),
			Categories:  []string{"men's clothing"}, // Категория мужской одежды
		}
		res = user.createListing(t, clothing)
		require.Equal(t, http.StatusOK, res.StatusCode)
		
		// Получаем ID созданного объявления одежды
		var clothingResp listing.CreateListingResponse
		err = json.NewDecoder(res.Body).Decode(&clothingResp)
		require.NoError(t, err)
		t.Logf("Создано объявление одежды с ID: %s, категории: %v", clothingResp.ID, clothingResp.Categories)

		furniture := listing.CreateListingRequest{
			Title:       "Диван угловой",
			Description: "Удобный угловой диван",
			Price:       25000,
			Currency:    models.Currency("RUB"),
			Categories:  []string{"furniture"}, // Категория мебели
		}
		res = user.createListing(t, furniture)
		require.Equal(t, http.StatusOK, res.StatusCode)
		
		// Получаем ID созданного объявления мебели
		var furnitureResp listing.CreateListingResponse
		err = json.NewDecoder(res.Body).Decode(&furnitureResp)
		require.NoError(t, err)
		t.Logf("Создано объявление мебели с ID: %s, категории: %v", furnitureResp.ID, furnitureResp.Categories)

		// Поиск по категории смартфонов
		req := getSearchListingsRequest("", 10, "", "relevance", "")
		req.CategoryID = "smartphones" // Используем одну категорию в запросе поиска
		t.Logf("Выполняем поиск по категории: %s", req.CategoryID)

		resp := user.searchListings(t, req)
		t.Logf("Найдено результатов: %d", len(resp.Results))

		// Проверяем, что найдено только объявление из категории смартфонов
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске по категории смартфонов")
		
		// Выводим информацию о каждом найденном объявлении
		for i, result := range resp.Results {
			t.Logf("Результат #%d: Заголовок=%s", 
				i+1, result.Title)
		}
		
		found := false
		for _, result := range resp.Results {
			if result.Title == "iPhone 13 Pro Max" {
				found = true
			}
			// Проверяем, что не найдены объявления из других категорий
			assert.NotEqual(t, "Мужская куртка", result.Title, "Найдено объявление из неправильной категории")
			assert.NotEqual(t, "Диван угловой", result.Title, "Найдено объявление из неправильной категории")
		}
		require.True(t, found, "Не найдено объявление из категории смартфонов")

		// Поиск по категории мужской одежды
		req = getSearchListingsRequest("", 10, "", "relevance", "")
		req.CategoryID = "men's clothing" // Используем одну категорию в запросе поиска

		resp = user.searchListings(t, req)

		// Проверяем, что найдено только объявление из категории мужской одежды
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске по категории мужской одежды")
		found = false
		for _, result := range resp.Results {
			if result.Title == "Мужская куртка" {
				found = true
			}
			// Проверяем, что не найдены объявления из других категорий
			assert.NotEqual(t, "iPhone 13 Pro Max", result.Title, "Найдено объявление из неправильной категории")
			assert.NotEqual(t, "Диван угловой", result.Title, "Найдено объявление из неправильной категории")
		}
		require.True(t, found, "Не найдено объявление из категории мужской одежды")

		// Поиск по категории мебели
		req = getSearchListingsRequest("", 10, "", "relevance", "")
		req.CategoryID = "furniture" // Используем одну категорию в запросе поиска

		resp = user.searchListings(t, req)

		// Проверяем, что найдено только объявление из категории мебели
		require.NotEmpty(t, resp.Results, "Ничего не найдено при поиске по категории мебели")
		found = false
		for _, result := range resp.Results {
			if result.Title == "Диван угловой" {
				found = true
			}
			// Проверяем, что не найдены объявления из других категорий
			assert.NotEqual(t, "iPhone 13 Pro Max", result.Title, "Найдено объявление из неправильной категории")
			assert.NotEqual(t, "Мужская куртка", result.Title, "Найдено объявление из неправильной категории")
		}
		require.True(t, found, "Не найдено объявление из категории мебели")

		// Поиск по несуществующей категории
		req = getSearchListingsRequest("", 10, "", "relevance", "")
		req.CategoryID = "nonexistent_category"

		resp = user.searchListings(t, req)

		// Проверяем, что ничего не найдено
		require.Empty(t, resp.Results, "Найдены объявления при поиске по несуществующей категории")
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
		Location: &models.Location{
			ID:   "moscow_center",
			Name: "Красная площадь",
			Area: models.Area{
				Coordinates: models.Coordinates{
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
				Coordinates: models.Coordinates{
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
				Coordinates: models.Coordinates{
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
				Coordinates: models.Coordinates{
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
