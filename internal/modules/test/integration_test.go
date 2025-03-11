package modules

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

	gorm, pool, err := db.NewSqlDB(context.Background(), cfg.DB.Url, lg.Logger, false)
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

func TestCreateListing(t *testing.T) {
	// Инициализация тестовой БД и роутера
	app := createTestApp(t)
	user := app.createUser(t)
	t.Run("Success create listing", func(t *testing.T) {
		listingInput := listing.CreateListingRequest{
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
