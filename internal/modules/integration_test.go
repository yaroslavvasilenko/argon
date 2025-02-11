package iphone

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/controller"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/service"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"github.com/yaroslavvasilenko/argon/internal/router"
)

type TestApp struct {
	fiber        *fiber.App
	listingStore *storage.Storage
	pool         *pgxpool.Pool
}

type user struct {
	TestApp
}

func (app *TestApp) createUser(t *testing.T) *user {
	return &user{
		TestApp: *app,
	}
}

func createTestApp(t *testing.T) *TestApp {
	// Init configuration
	config.LoadConfig()

	cfg := config.GetConfig()

	lg := logger.NewLogger(cfg)

	gorm, pool, err := db.NewSqlDB(context.Background(), cfg.DB.Url, lg.Logger, true)
	require.NoError(t, err)

	// Migrate
	err = database.Migrate(cfg.DB.Url)
	require.NoError(t, err)

	storagesDB := storage.NewStorage(gorm, pool)

	service := service.NewService(storagesDB, pool, lg)
	controller := controller.NewHandler(service)
	// init router
	r := router.NewApiRouter(controller)

	app := &TestApp{
		fiber: r,
		pool:  pool,
	}

	return app
}

func TestCreateListing(t *testing.T) {
	// Инициализация тестовой БД и роутера
	app := createTestApp(t)
	user := app.createUser(t)
	t.Run("Success create listing", func(t *testing.T) {
		listingInput := models.Listing{
			Title:       "Тестовая квартира",
			Description: "Просторная квартира в центре",
			Price:       1000000,
			Currency:    models.RUB,
		}

		resp := user.createListing(t, listingInput)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		listOut := models.Listing{}
		json.NewDecoder(resp.Body).Decode(&listOut)

		// Проверка ответа
		assert.NotZero(t, listOut.ID)
		assert.Equal(t, listingInput.Title, listOut.Title)
		assert.Equal(t, listingInput.Description, listOut.Description)
		assert.Equal(t, listingInput.Price, listOut.Price)
		assert.Equal(t, listingInput.Currency, listOut.Currency)

		// Проверка времени в UTC
		// now := time.Now().UTC()
		// assert.WithinDuration(t, now, listOut.CreatedAt.UTC(), 2*time.Second)
		// assert.WithinDuration(t, now, listOut.UpdatedAt.UTC(), 2*time.Second)

		assert.Empty(t, listOut.ViewsCount)
		assert.Nil(t, listOut.DeletedAt)
	})
}

func TestSearchListings(t *testing.T) {
	app := createTestApp(t)
	app.cleanDb(t)
	user := app.createUser(t)

	iphone1 := models.Listing{
		Title:       "iPhone 14 Pro",
		Description: "Новый iPhone 14 Pro, 256GB, цвет: космический черный",
		Price:       100001,
		Currency:    models.Currency("RUB"),
	}

	resp := user.createListing(t, iphone1)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	t.Run("Successful search 1 listing", func(t *testing.T) {
		// Выполняем поиск объявления
		req := getSearchListingsRequest("iPhone", 10, "", "relevance_desc", "")

		resp := user.searchListings(t, req)

		require.Len(t, resp.Results, 1)
		foundListing := resp.Results[0]
		assert.Equal(t, iphone1.Title, foundListing.Title)
		assert.Equal(t, iphone1.Description, foundListing.Description)
		assert.Equal(t, iphone1.Price, foundListing.Price)
		assert.Equal(t, iphone1.Currency, foundListing.Currency)
	})

	iphone2 := models.Listing{
		Title:       "iPhone 15 Pro",
		Description: "Новый iPhone 15 Pro, 256GB, цвет: космический черный",
		Price:       100002,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone2)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone3 := models.Listing{
		Title:       "iPhone 16 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100003,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone3)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone4 := models.Listing{
		Title:       "iPhone 17 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100004,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone4)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone5 := models.Listing{
		Title:       "iPhone 18 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100005,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone5)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone6 := models.Listing{
		Title:       "iPhone 19 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100006,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone6)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone7 := models.Listing{
		Title:       "iPhone 20 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100007,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone7)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone8 := models.Listing{
		Title:       "iPhone 21 Pro",
		Description: "Новый iPhone 16 Pro, 256GB, цвет: космический черный",
		Price:       100008,
		Currency:    models.Currency("RUB"),
	}
	resp = user.createListing(t, iphone8)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	iphone9 := models.Listing{
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

	})

}

func getSearchListingsRequest(query string, limit int, cursor string, sortOrder string, searchID string) listing.SearchListingsRequest {
	return listing.SearchListingsRequest{
		Query:     query,
		Limit:     limit,
		Cursor:    cursor,
		SortOrder: sortOrder,
		SearchID:  searchID,
	}
}

func (user *user) searchListings(t *testing.T, req listing.SearchListingsRequest) listing.SearchListingsResponse {
	t.Helper()

	body, err := json.Marshal(req)
	require.NoError(t, err)

	reqSearch := httptest.NewRequest("POST", "/api/v1/search", bytes.NewReader(body)).WithContext(context.Background())
	reqSearch.Header.Set("Content-Type", "application/json")
	resp, err := user.fiber.Test(reqSearch, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var searchResp listing.SearchListingsResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	require.NoError(t, err)

	return searchResp
}

func (user *user) createListing(t *testing.T, l models.Listing) *http.Response {
	body, err := json.Marshal(l)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/v1/listing", bytes.NewReader(body)).WithContext(context.Background())
	req.Header.Set("Content-Type", "application/json")

	resp, err := user.fiber.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func (app *TestApp) cleanDb(t *testing.T) {
	ctx := context.Background()
	// Очищаем все таблицы
	_, err := app.pool.Exec(ctx, `
		TRUNCATE TABLE listings_search_en, listings_search_ru, listings_search_es, listings CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}
}
