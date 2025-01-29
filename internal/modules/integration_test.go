package iphone

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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
	t.Run("Успешное создание объявления", func(t *testing.T) {
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

	t.Run("Успешный поиск объявления", func(t *testing.T) {
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

	t.Run("Успешный поиск объявлений", func(t *testing.T) {
		// Выполняем поиск объявления
		req := getSearchListingsRequest("iPhone", 2, "", "price_asc", "")

		resp := user.searchListings(t, req)

		require.Len(t, resp.Results, 2)
		assert.Equal(t, iphone1.Title, resp.Results[0].Title)
		assert.Equal(t, iphone2.Title, resp.Results[1].Title)

		req = getSearchListingsRequest("iPhone", 10, resp.CursorAfter, "price_asc", resp.SearchID)

		resp = user.searchListings(t, req)
		require.Len(t, resp.Results, 1)
		assert.Equal(t, iphone3.Title, resp.Results[0].Title)
	})
}

func getSearchListingsRequest(query string, limit int, cursor string, sortOrder string, searchID string) listing.SearchListingsRequest {
	return listing.SearchListingsRequest{
		Query: query,
		Limit: limit,
		Cursor: cursor,
		SortOrder: sortOrder,
		SearchID: searchID,
	}
}

func (user *user) searchListings(t *testing.T, req listing.SearchListingsRequest) listing.SearchListingsResponse {
	t.Helper()
	baseURL := "/api/v1/search"
	params := url.Values{}

	if req.Query != "" {
		params.Add("query", req.Query)
	}
	if req.Limit > 0 {
		params.Add("limit", strconv.Itoa(req.Limit))
	}
	if req.Cursor != "" {
		params.Add("cursor", req.Cursor)
	}
	if req.SortOrder != "" {
		params.Add("sort_order", req.SortOrder)
	}
	if req.SearchID != "" {
		params.Add("search_id", req.SearchID)
	}
	if req.Category != "" {
		params.Add("category", req.Category)
	}

	url := baseURL
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	reqSearch := httptest.NewRequest("GET", url, nil).WithContext(context.Background())
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

	resp, err := user.fiber.Test(req)
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
