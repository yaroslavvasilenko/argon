package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	bstorage "github.com/yaroslavvasilenko/argon/internal/modules/boost/storage"
	"gorm.io/gorm"
)

type Listing struct {
	gorm  *gorm.DB
	pool  *pgxpool.Pool
	boost *bstorage.Boost
}

func NewListing(db *gorm.DB, pool *pgxpool.Pool, boost *bstorage.Boost) *Listing {
	return &Listing{gorm: db, pool: pool, boost: boost}
}

// ListingDetails содержит все данные для создания объявления
type ListingDetails struct {
	Listing         models.Listing
	Categories      []string
	Location        models.Location
	Characteristics map[string]interface{}
}

// BatchCreateListingsWithDetails создает несколько объявлений с их категориями, локациями и характеристиками
func (s *Listing) BatchCreateListingsWithDetails(ctx context.Context, listingsDetails []ListingDetails) error {
	// Начинаем общую транзакцию для всех операций
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Для каждого объявления создаем его и связанные с ним данные
	for _, details := range listingsDetails {
		// 1. Вставляем основные данные листинга
		_, err = tx.Exec(ctx, `
			INSERT INTO listings (
				id, 
				title, 
				original_description, 
				created_at, 
				updated_at, 
				deleted_at,
				price,
				views_count,
				currency
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			details.Listing.ID,
			details.Listing.Title,
			details.Listing.Description,
			details.Listing.CreatedAt,
			details.Listing.UpdatedAt,
			details.Listing.DeletedAt,
			details.Listing.Price,
			details.Listing.ViewsCount,
			details.Listing.Currency,
		)
		if err != nil {
			return err
		}

		// 2. Вставляем информацию о локации, если она предоставлена
		if details.Location.ID != "" {
			_, err = tx.Exec(ctx, `
				INSERT INTO locations (
					id,
					listing_id,
					name,
					latitude,
					longitude,
					radius
				) VALUES ($1, $2, $3, $4, $5, $6)
			`,
				details.Location.ID,
				details.Listing.ID,
				details.Location.Name,
				details.Location.Area.Coordinates.Lat,
				details.Location.Area.Coordinates.Lng,
				int32(details.Location.Area.Radius),
			)
			if err != nil {
				return err
			}
		}

		// 3. Вставляем категории
		if len(details.Categories) > 0 {
			// Подготавливаем batch для массовой вставки категорий
			batch := &pgx.Batch{}
			for _, category := range details.Categories {
				batch.Queue(`
					INSERT INTO listing_categories (listing_id, category_id)
					VALUES ($1, $2)
				`, details.Listing.ID, category)
			}

			// Выполняем batch запрос
			br := tx.SendBatch(ctx, batch)

			// Проверяем результаты каждой операции в batch
			for i := 0; i < batch.Len(); i++ {
				_, err := br.Exec()
				if err != nil {
					brCloseErr := br.Close()
					if brCloseErr != nil {
						return errors.New(err.Error() + "; также ошибка при закрытии batch: " + brCloseErr.Error())
					}
					return err
				}
			}

			// Закрываем batch
			if err := br.Close(); err != nil {
				return err
			}
		}

		// 4. Вставляем характеристики, если они предоставлены
		if details.Characteristics != nil && len(details.Characteristics) > 0 {
			// Преобразуем map в JSON
			characteristicsJSON, err := json.Marshal(details.Characteristics)
			if err != nil {
				return err
			}

			// Вставляем характеристики в таблицу
			_, err = tx.Exec(ctx, `
				INSERT INTO listing_characteristics (
					listing_id,
					characteristics
				) VALUES ($1, $2)
			`,
				details.Listing.ID,
				characteristicsJSON,
			)
			if err != nil {
				return err
			}
		}
	}

	// Если все операции успешны, фиксируем транзакцию
	return tx.Commit(ctx)
}

func (s *Listing) CreateListing(ctx context.Context, listing models.Listing, categories []string, location models.Location, characteristics map[string]interface{}) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Вставляем основные данные листинга
	_, err = tx.Exec(ctx, `
		INSERT INTO listings (
			id, 
			title, 
			original_description, 
			created_at, 
			updated_at, 
			deleted_at,
			price,
			views_count,
			currency
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		listing.ID,
		listing.Title,
		listing.Description,
		listing.CreatedAt,
		listing.UpdatedAt,
		listing.DeletedAt,
		listing.Price,
		listing.ViewsCount,
		listing.Currency,
	)
	if err != nil {
		return err
	}

	// Вставляем информацию о локации, если она предоставлена
	var locationID sql.NullString
	var locationName sql.NullString
	var latitude, longitude sql.NullFloat64
	var radius sql.NullInt32

	// Проверяем, что ID локации не пустой
	if location.ID != "" {
		locationID.String = location.ID
		locationID.Valid = true

		locationName.String = location.Name
		locationName.Valid = true

		latitude.Float64 = location.Area.Coordinates.Lat
		latitude.Valid = true

		longitude.Float64 = location.Area.Coordinates.Lng
		longitude.Valid = true

		radius.Int32 = int32(location.Area.Radius)
		radius.Valid = true

		_, err = tx.Exec(ctx, `
			INSERT INTO locations (
				id,
				listing_id,
				name,
				latitude,
				longitude,
				radius
			) VALUES ($1, $2, $3, $4, $5, $6)
		`,
			locationID.String,
			listing.ID,
			locationName.String,
			latitude.Float64,
			longitude.Float64,
			radius.Int32,
		)
		if err != nil {
			return err
		}
	}

	// Вставляем категории
	if len(categories) > 0 {
		// Подготавливаем batch для массовой вставки категорий
		batch := &pgx.Batch{}
		for _, category := range categories {
			batch.Queue(`
				INSERT INTO listing_categories (listing_id, category_id)
				VALUES ($1, $2)
			`, listing.ID, category)
		}

		// Выполняем batch запрос
		br := tx.SendBatch(ctx, batch)

		// Проверяем результаты каждой операции в batch
		for i := 0; i < batch.Len(); i++ {
			_, err := br.Exec()
			if err != nil {
				// Закрываем batch перед возвратом ошибки
				brCloseErr := br.Close()
				if brCloseErr != nil {
					// Логируем ошибку закрытия, но возвращаем основную ошибку
					// В реальном коде здесь можно использовать логгер
					return errors.New(err.Error() + "; также ошибка при закрытии batch: " + brCloseErr.Error())
				}
				return err
			}
		}

		// Закрываем batch перед коммитом транзакции
		if err := br.Close(); err != nil {
			return err
		}
	}

	// Вставляем характеристики, если они предоставлены
	if characteristics != nil && len(characteristics) > 0 {
		// Логируем характеристики для отладки
		fmt.Printf("Характеристики для сохранения: %+v\n", characteristics)
		
		// Преобразуем map в JSON
		characteristicsJSON, err := json.Marshal(characteristics)
		if err != nil {
			return err
		}
		
		// Логируем JSON для отладки
		fmt.Printf("JSON характеристик: %s\n", string(characteristicsJSON))

		// Вставляем характеристики в таблицу
		_, err = tx.Exec(ctx, `
			INSERT INTO listing_characteristics (
				listing_id,
				characteristics
			) VALUES ($1, $2)
		`,
			listing.ID,
			characteristicsJSON,
		)
		if err != nil {
			return err
		}
	}

	// Если все операции успешны, фиксируем транзакцию
	return tx.Commit(ctx)
}

func (s *Listing) GetListing(ctx context.Context, pID string) (models.Listing, error) {
	listing := models.Listing{}

	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		First(&listing).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Listing{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return models.Listing{}, err
	}

	return listing, nil
}

type FullListing struct {
	Listing         models.Listing
	Categories      models.Category
	Location        models.Location
	Characteristics map[string]interface{}
	Boosts          []models.Boost
}

func (s *Listing) GetFullListing(ctx context.Context, pID string) (FullListing, error) {
	resp := FullListing{}

	// Проверяем корректность UUID
	listingID, err := uuid.Parse(pID)
	if err != nil {
		return resp, fiber.NewError(fiber.StatusBadRequest, "Некорректный ID объявления")
	}

	// Используем один запрос с LEFT JOIN для получения всех данных
	query := `
		WITH categories AS (
			SELECT 
				listing_id, 
				array_agg(category_id) AS category_ids
			FROM listing_categories
			WHERE listing_id = $1
			GROUP BY listing_id
		)
		SELECT 
			l.id, 
			l.title, 
			l.original_description, 
			l.created_at, 
			l.updated_at, 
			l.deleted_at,
			l.price,
			l.views_count,
			l.currency,
			c.category_ids,
			loc.id,
			loc.name,
			loc.latitude,
			loc.longitude,
			loc.radius
		FROM listings l
		LEFT JOIN categories c ON l.id = c.listing_id
		LEFT JOIN locations loc ON l.id = loc.listing_id
		WHERE l.id = $1 AND l.deleted_at IS NULL
	`

	row := s.pool.QueryRow(ctx, query, listingID)

	var listing models.Listing
	var deletedAt sql.NullTime
	var currencyStr string
	var categoryIDs []string

	// Для локации используем Nullable-типы, так как данные могут отсутствовать из-за LEFT JOIN
	var locationID sql.NullString
	var locationName sql.NullString
	var latitude, longitude sql.NullFloat64
	var radius sql.NullInt32

	err = row.Scan(
		&listing.ID,
		&listing.Title,
		&listing.Description,
		&listing.CreatedAt,
		&listing.UpdatedAt,
		&deletedAt,
		&listing.Price,
		&listing.ViewsCount,
		&currencyStr,
		&categoryIDs,
		&locationID,
		&locationName,
		&latitude,
		&longitude,
		&radius,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, fiber.NewError(fiber.StatusNotFound, "Объявление не найдено")
		}
		return resp, err
	}

	// Заполняем основную информацию о листинге
	if deletedAt.Valid {
		listing.DeletedAt = &deletedAt.Time
	}
	listing.Currency = models.Currency(currencyStr)
	resp.Listing = listing

	// Заполняем информацию о категориях
	resp.Categories = models.Category{
		ID:        categoryIDs,
		ListingID: listingID.String(),
	}

	// Заполняем информацию о локации, если она есть
	if locationID.Valid {
		var location models.Location
		location.ID = locationID.String
		location.ListingID = listingID
		location.Name = locationName.String
		location.Area = models.Area{
			Radius: int(radius.Int32),
		}
		location.Area.Coordinates.Lat = latitude.Float64
		location.Area.Coordinates.Lng = longitude.Float64
		resp.Location = location
	}

	// Получаем характеристики объявления
	characteristics, err := s.GetListingCharacteristics(ctx, listingID)
	if err == nil { // Игнорируем ошибку, так как характеристики могут отсутствовать
		resp.Characteristics = characteristics
	} else {
		// Инициализируем пустую карту, чтобы избежать nil в ответе
		resp.Characteristics = make(map[string]interface{})
	}

	boost, err := s.boost.GetBoosts(ctx, listingID)
	if err != nil {
		return resp, err
	}
	resp.Boosts = boost

	return resp, nil
}

func (s *Listing) DeleteListing(ctx context.Context, pID string) error {
	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Listing) UpdateFullListing(ctx context.Context, listing models.Listing, categories []string, location models.Location, characteristics map[string]interface{}) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Обновляем основные данные листинга
	_, err = tx.Exec(ctx, `
		UPDATE listings 
		SET 
			title = $1, 
			original_description = $2, 
			updated_at = $3, 
			price = $4,
			currency = $5
		WHERE id = $6 AND deleted_at IS NULL
	`,
		listing.Title,
		listing.Description,
		listing.UpdatedAt,
		listing.Price,
		listing.Currency,
		listing.ID,
	)
	if err != nil {
		return err
	}

	// Удаляем существующие категории для этого листинга
	_, err = tx.Exec(ctx, `DELETE FROM listing_categories WHERE listing_id = $1`, listing.ID)
	if err != nil {
		return err
	}

	// Вставляем новые категории
	if len(categories) > 0 {
		batch := &pgx.Batch{}
		for _, category := range categories {
			batch.Queue(`
				INSERT INTO listing_categories (listing_id, category_id)
				VALUES ($1, $2)
			`, listing.ID, category)
		}

		br := tx.SendBatch(ctx, batch)

		for i := 0; i < batch.Len(); i++ {
			_, err := br.Exec()
			if err != nil {
				brCloseErr := br.Close()
				if brCloseErr != nil {
					return errors.New(err.Error() + "; также ошибка при закрытии batch: " + brCloseErr.Error())
				}
				return err
			}
		}

		if err := br.Close(); err != nil {
			return err
		}
	}

	// Обновляем информацию о локации
	if location.ID != "" {
		// Проверяем, существует ли уже локация для этого листинга
		var locationExists bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM locations WHERE listing_id = $1)`, listing.ID).Scan(&locationExists)
		if err != nil {
			return err
		}

		if locationExists {
			// Обновляем существующую локацию
			_, err = tx.Exec(ctx, `
				UPDATE locations 
				SET 
					id = $1,
					name = $2,
					latitude = $3,
					longitude = $4,
					radius = $5
				WHERE listing_id = $6
			`,
				location.ID,
				location.Name,
				location.Area.Coordinates.Lat,
				location.Area.Coordinates.Lng,
				int32(location.Area.Radius),
				listing.ID,
			)
		} else {
			// Вставляем новую локацию
			_, err = tx.Exec(ctx, `
				INSERT INTO locations (
					id,
					listing_id,
					name,
					latitude,
					longitude,
					radius
				) VALUES ($1, $2, $3, $4, $5, $6)
			`,
				location.ID,
				listing.ID,
				location.Name,
				location.Area.Coordinates.Lat,
				location.Area.Coordinates.Lng,
				location.Area.Radius,
			)
		}

		if err != nil {
			return err
		}
	} else {
		// Если локация не предоставлена, удаляем существующую (если есть)
		_, err = tx.Exec(ctx, `DELETE FROM locations WHERE listing_id = $1`, listing.ID)
		if err != nil {
			return err
		}
	}

	// Обновляем характеристики объявления
	if characteristics != nil && len(characteristics) > 0 {
		// Преобразуем map в JSON
		characteristicsJSON, err := json.Marshal(characteristics)
		if err != nil {
			return err
		}

		// Проверяем, существуют ли уже характеристики для этого объявления
		var characteristicsExists bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM listing_characteristics WHERE listing_id = $1)`, listing.ID).Scan(&characteristicsExists)
		if err != nil {
			return err
		}

		if characteristicsExists {
			// Обновляем существующие характеристики
			_, err = tx.Exec(ctx, `
				UPDATE listing_characteristics 
				SET characteristics = $1 
				WHERE listing_id = $2
			`, characteristicsJSON, listing.ID)
		} else {
			// Вставляем новые характеристики
			_, err = tx.Exec(ctx, `
				INSERT INTO listing_characteristics (listing_id, characteristics) 
				VALUES ($1, $2)
			`, listing.ID, characteristicsJSON)
		}

		if err != nil {
			return err
		}
	} else {
		// Если характеристики не предоставлены, удаляем существующие (если есть)
		_, err = tx.Exec(ctx, `DELETE FROM listing_characteristics WHERE listing_id = $1`, listing.ID)
		if err != nil {
			return err
		}
	}

	// Фиксируем транзакцию
	return tx.Commit(ctx)
}

// GetListingCharacteristics получает характеристики объявления по его ID
func (s *Listing) GetListingCharacteristics(ctx context.Context, listingID uuid.UUID) (map[string]interface{}, error) {
	var characteristicsJSON []byte

	// Запрос для получения характеристик объявления
	err := s.pool.QueryRow(ctx, `
		SELECT characteristics 
		FROM listing_characteristics 
		WHERE listing_id = $1
	`, listingID).Scan(&characteristicsJSON)

	if err != nil {
		return nil, err
	}

	// Если характеристики не найдены, возвращаем пустую карту
	if characteristicsJSON == nil {
		return make(map[string]interface{}), nil
	}

	// Распаковываем JSON в карту
	var characteristics map[string]interface{}
	if err := json.Unmarshal(characteristicsJSON, &characteristics); err != nil {
		return nil, err
	}

	return characteristics, nil
}

func (s *Listing) GetCharacteristicValues(ctx context.Context, characteristicKeys []string) (models.Filters, error) {
	// Создаем результирующую карту для хранения значений характеристик
	result := make(models.Filters)

	// Для каждого ключа характеристики выполняем отдельный запрос
	for _, key := range characteristicKeys {
		switch key {
		case "price":
			// Для цены получаем минимальное и максимальное значение
			var minPrice, maxPrice *float64
			query := `SELECT MIN(price), MAX(price) FROM listings WHERE deleted_at IS NULL`
			err := s.pool.QueryRow(ctx, query).Scan(&minPrice, &maxPrice)
			if err != nil {
				return nil, fmt.Errorf("ошибка при получении диапазона цен: %w", err)
			}
			
			// Устанавливаем значения по умолчанию, если в базе нет данных
			min, max := 0, 0
			if minPrice != nil {
				min = int(*minPrice)
			}
			if maxPrice != nil {
				max = int(*maxPrice)
			}
			
			result[key] = models.PriceFilter{
				Min: min,
				Max: max,
			}

		case "brand", "condition", "color", "season":
			// Для строковых характеристик получаем уникальные значения
			query := `
				SELECT DISTINCT jsonb_array_elements_text(characteristics->$1) AS value
				FROM listing_characteristics
				WHERE characteristics ? $1
				ORDER BY value
			`
			rows, err := s.pool.Query(ctx, query, key)
			if err != nil {
				return nil, fmt.Errorf("ошибка при получении уникальных значений для %s: %w", key, err)
			}
			defer rows.Close()

			values := make([]string, 0)
			for rows.Next() {
				var value string
				if err := rows.Scan(&value); err != nil {
					return nil, fmt.Errorf("ошибка при сканировании значения для %s: %w", key, err)
				}
				values = append(values, value)
			}
			
			if key == "color" {
				result[key] = models.ColorFilter(values)
			} else {
				result[key] = models.DropdownFilter(values)
			}

		case "stocked":
			// Для булевых характеристик просто возвращаем true/false
			result[key] = models.CheckboxFilter(false)

		case "height", "width", "depth", "weight", "area", "volume":
			// Для числовых характеристик получаем минимальное и максимальное значение
			query := `
				SELECT 
					MIN(CAST(characteristics->$1 AS NUMERIC)), 
					MAX(CAST(characteristics->$1 AS NUMERIC))
				FROM listing_characteristics
				WHERE characteristics ? $1
			`
			var minValue, maxValue *float64
			err := s.pool.QueryRow(ctx, query, key).Scan(&minValue, &maxValue)
			if err != nil {
				// Если нет данных, устанавливаем значения по умолчанию
				if err == pgx.ErrNoRows {
					result[key] = models.DimensionFilter{
						Min:       0,
						Max:       0,
						Dimension: "",
					}
					continue
				}
				return nil, fmt.Errorf("ошибка при получении диапазона для %s: %w", key, err)
			}
			
			// Устанавливаем значения по умолчанию, если в базе нет данных
			min, max := 0, 0
			if minValue != nil {
				min = int(*minValue)
			}
			if maxValue != nil {
				max = int(*maxValue)
			}
			
			result[key] = models.DimensionFilter{
				Min:       min,
				Max:       max,
				Dimension: getDimensionUnit(key),
			}

		default:
			// Для неизвестных характеристик пропускаем
			continue
		}
	}

	return result, nil
}

// getDimensionUnit возвращает единицу измерения для указанной характеристики
func getDimensionUnit(key string) string {
	switch key {
	case "height", "width", "depth":
		return "см"
	case "weight":
		return "кг"
	case "area":
		return "м²"
	case "volume":
		return "л"
	default:
		return ""
	}
}