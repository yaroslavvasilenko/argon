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
				images,
				price,
				views_count,
				currency,
				created_at, 
				updated_at, 
				deleted_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			details.Listing.ID,
			details.Listing.Title,
			details.Listing.Description,
			details.Listing.Images,
			details.Listing.Price,
			details.Listing.ViewsCount,
			details.Listing.Currency,
			details.Listing.CreatedAt,
			details.Listing.UpdatedAt,
			details.Listing.DeletedAt,
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
			images,
			price,
			views_count,
			currency,
			created_at, 
			updated_at, 
			deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		listing.ID,
		listing.Title,
		listing.Description,
		listing.Images,
		listing.Price,
		listing.ViewsCount,
		listing.Currency,
		listing.CreatedAt,
		listing.UpdatedAt,
		listing.DeletedAt,
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
			l.images,
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
		&listing.Images,
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
			images = $3,
			updated_at = $4, 
			price = $5,
			currency = $6
		WHERE id = $7 AND deleted_at IS NULL
	`,
		listing.Title,
		listing.Description,
		listing.Images,
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

// GetCategoryFilters получает все доступные фильтры для указанной категории
func (s *Listing) GetCategoryFilters(ctx context.Context, categoryID string) (models.Filters, error) {
	// Создаем результирующую карту для хранения фильтров
	result := make(models.Filters)

	// Сначала проверим, существует ли категория
	var categoryExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM listing_categories WHERE category_id = $1)`
	err := s.pool.QueryRow(ctx, checkQuery, categoryID).Scan(&categoryExists)
	if err != nil {
		return nil, fmt.Errorf("ошибка при проверке существования категории %s: %w", categoryID, err)
	}

	// Если категория не существует или нет товаров в этой категории, вернем пустые фильтры
	if !categoryExists {
		return result, nil
	}

	// SQL-запрос для получения минимальной и максимальной цены, а также всех характеристик в категории
	query := `
	WITH category_listings AS (
		SELECT l.id, l.price
		FROM listings l
		JOIN listing_categories lc ON l.id = lc.listing_id
		WHERE lc.category_id = $1
		AND l.deleted_at IS NULL
	)
	SELECT 
		MIN(cl.price) AS min_price,
		MAX(cl.price) AS max_price,
		(
			-- Подзапрос для получения всех уникальных характеристик в категории
			SELECT jsonb_object_agg(
				key, 
				CASE
					-- Для цены возвращаем объект с min и max значениями
					WHEN key = 'price' THEN jsonb_build_object(
						'min', MIN(cl.price),
						'max', MAX(cl.price)
					)
					-- Для цветов возвращаем массив уникальных значений
					WHEN key = 'color' THEN (
						SELECT jsonb_agg(DISTINCT value)
						FROM (
							-- Обработка массивов
							SELECT jsonb_array_elements_text(lch.characteristics->'color') AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'color'
							AND jsonb_typeof(lch.characteristics->'color') = 'array'
							UNION ALL
							-- Обработка скалярных значений
							SELECT lch.characteristics->>'color' AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'color'
							AND jsonb_typeof(lch.characteristics->'color') != 'array'
						) subq
						WHERE value IS NOT NULL
					)
					-- Для брендов возвращаем массив уникальных значений
					WHEN key = 'brand' THEN (
						SELECT jsonb_agg(DISTINCT value)
						FROM (
							-- Обработка массивов
							SELECT jsonb_array_elements_text(lch.characteristics->'brand') AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'brand'
							AND jsonb_typeof(lch.characteristics->'brand') = 'array'
							UNION ALL
							-- Обработка скалярных значений
							SELECT lch.characteristics->>'brand' AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'brand'
							AND jsonb_typeof(lch.characteristics->'brand') != 'array'
						) subq
						WHERE value IS NOT NULL
					)
					-- Для состояния (новый/б/у) возвращаем массив уникальных значений
					WHEN key = 'condition' THEN (
						SELECT jsonb_agg(DISTINCT value)
						FROM (
							-- Обработка массивов
							SELECT jsonb_array_elements_text(lch.characteristics->'condition') AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'condition'
							AND jsonb_typeof(lch.characteristics->'condition') = 'array'
							UNION ALL
							-- Обработка скалярных значений
							SELECT lch.characteristics->>'condition' AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'condition'
							AND jsonb_typeof(lch.characteristics->'condition') != 'array'
						) subq
						WHERE value IS NOT NULL
					)
					-- Для сезона возвращаем массив уникальных значений
					WHEN key = 'season' THEN (
						SELECT jsonb_agg(DISTINCT value)
						FROM (
							-- Обработка массивов
							SELECT jsonb_array_elements_text(lch.characteristics->'season') AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'season'
							AND jsonb_typeof(lch.characteristics->'season') = 'array'
							UNION ALL
							-- Обработка скалярных значений
							SELECT lch.characteristics->>'season' AS value
							FROM listing_characteristics lch
							JOIN listing_categories lc ON lch.listing_id = lc.listing_id
							WHERE lc.category_id = $1
							AND lch.characteristics ? 'season'
							AND jsonb_typeof(lch.characteristics->'season') != 'array'
						) subq
						WHERE value IS NOT NULL
					)
					-- Для булевых значений (например, "в наличии")
					WHEN key = 'stocked' THEN (
						SELECT jsonb_agg(DISTINCT (lch.characteristics->>'stocked')::boolean)
						FROM listing_characteristics lch
						JOIN listing_categories lc ON lch.listing_id = lc.listing_id
						WHERE lc.category_id = $1
						AND lch.characteristics ? 'stocked'
					)
					-- Для размерных характеристик (высота, ширина и т.д.)
					WHEN key IN ('height', 'width', 'depth', 'weight', 'area', 'volume') THEN (
						SELECT jsonb_build_object(
							'min', MIN(
								CASE 
									WHEN jsonb_typeof(lch.characteristics->key) = 'number' THEN (lch.characteristics->>key)::numeric
									ELSE NULL
								END
							),
							'max', MAX(
								CASE 
									WHEN jsonb_typeof(lch.characteristics->key) = 'number' THEN (lch.characteristics->>key)::numeric
									ELSE NULL
								END
							),
							'dimension', CASE
								WHEN key = 'height' OR key = 'width' OR key = 'depth' THEN 'cm'
								WHEN key = 'weight' THEN 'kg'
								WHEN key = 'area' THEN 'm2'
								WHEN key = 'volume' THEN 'l'
								ELSE ''
							END
						)
						FROM listing_characteristics lch
						JOIN listing_categories lc ON lch.listing_id = lc.listing_id
						WHERE lc.category_id = $1
					)
					ELSE NULL
				END
			)
			FROM (
				-- Получаем все ключи характеристик, которые есть в категории
				SELECT DISTINCT key
				FROM (
					SELECT jsonb_object_keys(lch.characteristics) AS key
					FROM listing_characteristics lch
					JOIN listing_categories lc ON lch.listing_id = lc.listing_id
					WHERE lc.category_id = $1
				) subq
			) keys
		) AS characteristics
	FROM category_listings cl;
	`

	// Выполняем запрос
	var minPrice, maxPrice *float64
	var characteristicsJSON []byte

	err = s.pool.QueryRow(ctx, query, categoryID).Scan(&minPrice, &maxPrice, &characteristicsJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Если нет данных, возвращаем пустую карту фильтров
			return result, nil
		}
		return nil, fmt.Errorf("ошибка при получении фильтров для категории %s: %w", categoryID, err)
	}

	// Создаем фильтр цены
	if minPrice != nil && maxPrice != nil {
		result[models.CHAR_PRICE] = models.PriceFilter{
			Min: int(*minPrice),
			Max: int(*maxPrice),
		}
	}

	// Парсим JSON с характеристиками
	if len(characteristicsJSON) > 0 && characteristicsJSON != nil {
		var characteristics map[string]json.RawMessage
		if err := json.Unmarshal(characteristicsJSON, &characteristics); err != nil {
			// Если JSON пустой или null, просто вернем пустые фильтры
			if err.Error() == "unexpected end of JSON input" {
				return result, nil
			}
			return nil, fmt.Errorf("ошибка при разборе JSON характеристик: %w", err)
		}

		// Обрабатываем каждую характеристику
		for key, value := range characteristics {
			switch key {
			case models.CHAR_COLOR:
				// Для цвета
				var colors []string
				if err := json.Unmarshal(value, &colors); err != nil {
					continue // Пропускаем некорректные данные
				}
				if colors != nil && len(colors) > 0 {
					result[key] = models.ColorFilter{Options: colors}
				}

			case models.CHAR_BRAND, models.CHAR_CONDITION, models.CHAR_SEASON:
				// Для выпадающих списков
				var options []string
				if err := json.Unmarshal(value, &options); err != nil {
					continue // Пропускаем некорректные данные
				}
				if options != nil && len(options) > 0 {
					result[key] = models.DropdownFilter(options)
				}

			case models.CHAR_STOCKED:
				// Для булевых значений
				var boolValues []bool
				if err := json.Unmarshal(value, &boolValues); err != nil {
					continue // Пропускаем некорректные данные
				}
				if boolValues != nil && len(boolValues) > 0 {
					boolValue := boolValues[0]
					result[key] = models.CheckboxFilter(&boolValue)
				}

			case models.CHAR_HEIGHT, models.CHAR_WIDTH, models.CHAR_DEPTH, models.CHAR_WEIGHT, models.CHAR_AREA, models.CHAR_VOLUME:
				// Для размерных характеристик
				var dimensionFilter struct {
					Min       float64 `json:"min"`
					Max       float64 `json:"max"`
					Dimension string  `json:"dimension"`
				}
				if err := json.Unmarshal(value, &dimensionFilter); err != nil {
					continue // Пропускаем некорректные данные
				}
				
				result[key] = models.DimensionFilter{
					Min:       int(dimensionFilter.Min),
					Max:       int(dimensionFilter.Max),
					Dimension: dimensionFilter.Dimension,
				}
			}
		}
	}

	return result, nil
}

// getDimensionUnit возвращает единицу измерения для указанной характеристики
func getDimensionUnit(key string) string {
	switch key {
	case models.CHAR_HEIGHT, models.CHAR_WIDTH, models.CHAR_DEPTH:
		return models.CM
	case models.CHAR_WEIGHT:
		return models.KG
	case models.CHAR_AREA:
		return models.M2
	case models.CHAR_VOLUME:
		return models.L
	default:
		return ""
	}
}
