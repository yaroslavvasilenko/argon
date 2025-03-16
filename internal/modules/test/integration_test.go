package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stretchr/testify/require"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"github.com/yaroslavvasilenko/argon/internal/router"
)

type TestApp struct {
	fiber        *fiber.App
	listingStore *storage.Listing
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

	storages := modules.NewStorages(cfg, gorm, pool)
	services := modules.NewServices(storages, pool, lg)
	controller := modules.NewControllers(services)
	// init router
	r := router.NewApiRouter(controller)

	app := &TestApp{
		fiber: r,
		pool:  pool,
	}

	return app
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

func (user *user) createListing(t *testing.T, l listing.CreateListingRequest) *http.Response {
	body, err := json.Marshal(l)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "/api/v1/listing", bytes.NewReader(body)).WithContext(context.Background())
	req.Header.Set("Content-Type", "application/json")

	resp, err := user.fiber.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func (user *user) getCharacteristicsForCategory(t *testing.T, categoryIds []string, lang string) ([]models.CharacteristicItem, error) {
	req := struct {
		CategoryIds []string `json:"category_ids"`
	}{
		CategoryIds: categoryIds,
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq := httptest.NewRequest("POST", "/api/v1/categories/characteristics", bytes.NewReader(body)).WithContext(context.Background())
	httpReq.Header.Set("Content-Type", "application/json")

	// Установка языка, если он указан
	if lang != "" {
		httpReq.Header.Set(models.HeaderLanguage, lang)
	}

	resp, err := user.fiber.Test(httpReq, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var characteristics []models.CharacteristicItem
	err = json.NewDecoder(resp.Body).Decode(&characteristics)
	return characteristics, err
}

// getFiltersForCategory выполняет запрос к API для получения фильтров для указанной категории
func (user *user) getFiltersForCategory(t *testing.T, categoryId string, lang string) (models.Filters, error) {
	// Создаем URL с параметром запроса category_id
	url := fmt.Sprintf("/api/v1/categories/filters?category_id=%s", categoryId)

	httpReq := httptest.NewRequest("GET", url, nil).WithContext(context.Background())

	// Установка языка, если он указан
	if lang != "" {
		httpReq.Header.Set(models.HeaderLanguage, lang)
	}

	resp, err := user.fiber.Test(httpReq, -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var filters models.Filters
	err = json.NewDecoder(resp.Body).Decode(&filters)
	return filters, err
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

// func TestRunBenchmark(t *testing.T) {
// 	// Запускаем тест параллельно
// 	t.Parallel()

// 	// Определяем количество записей для генерации
// 	listingCount := 1000000 // Уменьшаем с 1 миллиона до 100 тысяч

// 	// Запускаем бенчмарк
// 	err := RunBenchmark(listingCount)
// 	require.NoError(t, err)
// }
