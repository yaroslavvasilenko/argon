package modules

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phuslu/log"

	"github.com/yaroslavvasilenko/argon/config"
	"github.com/yaroslavvasilenko/argon/database"
	"github.com/yaroslavvasilenko/argon/internal/core/db"
	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/yaroslavvasilenko/argon/internal/modules"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing/storage"
	"github.com/yaroslavvasilenko/argon/internal/router"
)

// BenchmarkApp представляет приложение для бенчмаркинга
type BenchmarkApp struct {
	fiber        *fiber.App
	listingStore *storage.Listing
	pool         *pgxpool.Pool
	rnd          *rand.Rand
}

func RunBenchmark(count int) error {
	// Отключаем вывод логов в консоль на время бенчмарка
	log.DefaultLogger.SetLevel(log.FatalLevel)
	
	app, err := NewBenchmarkApp()
	if err != nil {
		return fmt.Errorf("failed to create benchmark app: %w", err)
	}

	if err := app.GenerateListings(count); err != nil {
		return fmt.Errorf("failed to generate listings: %w", err)
	}
	app.Close()

	return nil
}

// NewBenchmarkApp создает новое приложение для бенчмаркинга
func NewBenchmarkApp() (*BenchmarkApp, error) {
	// Init configuration
	config.LoadConfig()

	cfg := config.GetConfig()

	lg := logger.NewLogger(cfg)
	
	// Инициализация генератора случайных чисел
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)

		// Создаем пустой логгер с уровнем логирования Fatal
	emptyLogger := log.Logger{}
	emptyLogger.SetLevel(log.FatalLevel)
	
	// Настраиваем пул соединений для больших нагрузок
	pgxConfig, err := pgxpool.ParseConfig(cfg.DB.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	
	// Увеличиваем максимальное количество соединений и таймауты
	pgxConfig.MaxConns = 50
	pgxConfig.MinConns = 10
	pgxConfig.MaxConnLifetime = 30 * time.Minute
	pgxConfig.MaxConnIdleTime = 15 * time.Minute
	pgxConfig.HealthCheckPeriod = 1 * time.Minute
	
	// Отключаем логирование SQL-запросов
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	
	// Создаем GORM с оптимизированными настройками
	gorm, _, err := db.NewSqlDB(context.Background(), cfg.DB.Url, emptyLogger, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Migrate
	err = database.Migrate(cfg.DB.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	storages := modules.NewStorages(cfg, gorm, pool, nil)
	services := modules.NewServices(storages, pool, lg)
	controller := modules.NewControllers(services)
	// init router
	r := router.NewApiRouter(controller)

	app := &BenchmarkApp{
		fiber:        r,
		listingStore: storages.Listing,
		pool:         pool,
		rnd:          rnd,
	}

	return app, nil
}

// CleanDB очищает базу данных перед бенчмарком
func (app *BenchmarkApp) CleanDB() error {
	ctx := context.Background()

	// Сначала очищаем таблицы
	_, err := app.pool.Exec(ctx, `
		TRUNCATE TABLE listings_search_en, listings_search_ru, listings_search_es, listings CASCADE;
	`)
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	// Затем освобождаем место на диске
	_, err = app.pool.Exec(ctx, `VACUUM FULL;`)
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	return nil
}

// Close закрывает соединение с базой данных
func (app *BenchmarkApp) Close() {
	if app.pool != nil {
		app.pool.Close()
	}
}

// BatchCreateListingsWithDetails создает объявления в базе данных в рамках транзакции
func (app *BenchmarkApp) BatchCreateListingsWithDetails(ctx context.Context, listingsDetails []storage.ListingDetails) error {
	// Создаем транзакцию для более эффективной вставки
	tx, err := app.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Отложенный rollback в случае ошибки
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()
	
	// Выполняем вставку в рамках транзакции
	err = app.listingStore.BatchCreateListingsWithDetails(ctx, listingsDetails)
	if err != nil {
		return fmt.Errorf("failed to create listings: %w", err)
	}
	
	// Фиксируем транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetDatabaseSize возвращает размер базы данных и таблиц
func (app *BenchmarkApp) GetDatabaseSize() (map[string]string, error) {
	var dbSize string
	tableSizes := make(map[string]string)

	// Получаем размер базы данных
	err := app.pool.QueryRow(context.Background(),
		"SELECT pg_size_pretty(pg_database_size('postgres')) as db_size",
	).Scan(&dbSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}
	tableSizes["database"] = dbSize

	// Получаем размеры всех таблиц
	rows, err := app.pool.Query(context.Background(), `
		SELECT
			relname as table_name,
			pg_size_pretty(pg_total_relation_size(relid)) as table_size
		FROM pg_catalog.pg_statio_user_tables
		ORDER BY pg_total_relation_size(relid) DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get table sizes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, tableSize string
		err := rows.Scan(&tableName, &tableSize)
		if err != nil {
			return nil, fmt.Errorf("failed to scan table size: %w", err)
		}
		tableSizes[tableName] = tableSize
	}

	return tableSizes, nil
}

// GenerateListings генерирует указанное количество объявлений в базе данных
func (app *BenchmarkApp) GenerateListings(count int) error {
	ctx := context.Background()
	startTime := time.Now()
	
	// Выводим только прогресс генерации объявлений
	fmt.Printf("\rГенерация %d объявлений...", count)

	// Размер батча для вставки
	const batchSize = 5000

	// Создаем локации, категории и характеристики заранее
	locations, err := app.generateLocations(10)
	if err != nil {
		return fmt.Errorf("failed to generate locations: %w", err)
	}

	categories := app.generateListingCategories(5)

	characteristics := []struct {
		role   string
		values []interface{}
	}{
		{"color", []interface{}{"white", "red", "blue", "green", "black"}},
		{"condition", []interface{}{"new", "used", "refurbished"}},
		{"brand", []interface{}{"Apple", "Samsung", "Sony", "LG", "Nike"}},
		{"stocked", []interface{}{true, false}},
		{"weight", []interface{}{"kg", "g", "t"}},
		{"area", []interface{}{"m2", "km2"}},
		{"volume", []interface{}{"l", "ml", "m3"}},
	}

	getRandomCharacteristic3 := func() map[string]interface{} {
		var result map[string]interface{}
		for _, characteristic := range characteristics {
			result = map[string]interface{}{
				"role":   characteristic.role,
				"values": characteristic.values,
			}

			if len(result) >= 3 {
				break
			}
		}
		return result
	}

	totalBatches := (count + batchSize - 1) / batchSize
	processedRecords := 0

	for batchNum := 0; batchNum < totalBatches; batchNum++ {

		// Определяем размер текущего батча
		currentBatchSize := batchSize
		if batchNum == totalBatches-1 {
			currentBatchSize = count - processedRecords
		}

		// Подготовка аргументов для текущего батча
		listingsDetails := make([]storage.ListingDetails, currentBatchSize)
		for i := 0; i < len(listingsDetails); i++ {
			id := uuid.New()
			title := app.generateTitle(processedRecords + i)
			description := app.generateDescription(processedRecords + i)
			price := app.generatePrice()
			viewsCount := app.rnd.Intn(1000)
			currency := app.generateCurrency()
			now := time.Now().UTC()
			locationID := locations[app.rnd.Intn(len(locations))]
			categoryID := categories[app.rnd.Intn(len(categories))]
			characteristicIDs := getRandomCharacteristic3()
			listingsDetails[i] = storage.ListingDetails{
				Listing: models.Listing{
					ID:           id,
					Title:        title,
					Description:  description,
					Price:        price,
					ViewsCount:   viewsCount,
					Currency:     currency,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
				Location: locationID,
				Categories: []string{categoryID},
				Characteristics: characteristicIDs,
			}
		}

		if err := app.BatchCreateListingsWithDetails(ctx, listingsDetails); err != nil {
			return fmt.Errorf("failed to create batch %d: %w", batchNum+1, err)
		}
			// Логирование прогресса
			elapsed := time.Since(startTime).Seconds()
			speed := float64(processedRecords) / elapsed
			estimatedTotal := float64(count) / speed
			estimatedRemaining := estimatedTotal - elapsed
			fmt.Printf("\rСгенерировано: %d/%d (батч %d/%d) | %.1f зап./сек. | Осталось: %.1f сек.   ", 
				processedRecords, count, batchNum+1, totalBatches, speed, estimatedRemaining)
	}

	// Выводим завершающее сообщение
	fmt.Printf("\rГотово! Сгенерировано %d объявлений за %.2f секунд.\n", count, time.Since(startTime).Seconds())

	return nil
}

// Вспомогательные функции для генерации данных

// Генерация заголовка объявления
func (app *BenchmarkApp) generateTitle(_ int) string {
	categories := []string{"Электроника", "Недвижимость", "Транспорт", "Бытовая техника"}
	category := categories[app.rnd.Intn(len(categories))]

	product := faker.Word()
	brand := faker.Word()

	return fmt.Sprintf("%s %s %s", brand, product, category)
}

// Генерация описания объявления
func (app *BenchmarkApp) generateDescription(_ int) string {
	condition := faker.Word()
	features := []string{
		faker.Word(),
		faker.Word(),
		faker.Word(),
	}

	description := fmt.Sprintf(
		"Продается товар в %s состоянии. Особенности: %s, %s, %s. "+
			"%s. %s. Местоположение: %s. %s",
		condition,
		features[0], features[1], features[2],
		faker.Sentence(),
		faker.Sentence(),
		faker.Name(),
		faker.Sentence(),
	)

	return description
}

// Генерация цены
func (app *BenchmarkApp) generatePrice() float64 {
	// Генерация более реалистичной цены с разными диапазонами
	priceRanges := []struct {
		min, max float64
		weight   int
	}{
		{1000, 10000, 40},    // Недорогие товары
		{10000, 100000, 35},  // Средний ценовой сегмент
		{100000, 500000, 20}, // Дорогие товары
		{500000, 2000000, 5}, // Премиум сегмент
	}

	// Выбор ценового диапазона на основе весов
	totalWeight := 0
	for _, r := range priceRanges {
		totalWeight += r.weight
	}

	rnd := app.rnd.Intn(totalWeight)
	currentWeight := 0

	for _, r := range priceRanges {
		currentWeight += r.weight
		if rnd < currentWeight {
			diff := r.max - r.min
			return r.min + (diff * app.rnd.Float64())
		}
	}

	return 1000.0 // Fallback
}

// Генерация валюты
func (app *BenchmarkApp) generateCurrency() models.Currency {
	currencies := []models.Currency{models.USD, models.EUR, models.RUB, models.ARS}
	return currencies[app.rnd.Intn(len(currencies))]
}

// generateLocations генерирует случайные локации
func (app *BenchmarkApp) generateLocations(count int) ([]models.Location, error) {
	locations := make([]models.Location, count)
	args := make([]interface{}, count*4)
	valuesPlaceholders := make([]string, count)
	now := time.Now().UTC()

	for i := 0; i < count; i++ {
		id := uuid.New()
		name := fmt.Sprintf("%s %s %s", faker.Word(), faker.Word(), faker.Word())
		offset := i * 4

		args[offset] = id
		args[offset+1] = name
		args[offset+2] = now
		args[offset+3] = now

		locations[i] = models.Location{
			ID:   id.String(),
			Name: name,
			Area: models.Area{
				Coordinates: models.Coordinates{
					Lat: app.rnd.Float64()*180 - 90,
					Lng: app.rnd.Float64()*360 - 180,
				},
				Radius: 1000,
			},
		}
		valuesPlaceholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4)
	}

	return locations, nil
}

// generateListingCategories генерирует случайные категории объявлений
func (app *BenchmarkApp) generateListingCategories(count int) []string {
	return []string{"electronics", "clothing", "furniture"}
}
