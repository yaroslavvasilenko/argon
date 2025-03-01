package storage

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
)

type Listing struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

func NewListing(db *gorm.DB, pool *pgxpool.Pool) *Listing {
	return &Listing{gorm: db, pool: pool}
}

func (s *Listing) CreateListing(ctx context.Context, listing models.Listing, categories []string, location models.Location) error {
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
	Listing    models.Listing
	Categories models.Category
	Location   models.Location
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

func (s *Listing) UpdateFullListing(ctx context.Context, listing models.Listing, categories []string, location models.Location) error {
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

	// Фиксируем транзакцию
	return tx.Commit(ctx)
}

func (s *Listing) SearchListingsByTitle(ctx context.Context, query string, limit int, cursorID *uuid.UUID, sort string) (*models.Listing, []models.Listing, error) {
	var (
		rows pgx.Rows
		err  error
	)

	// Если параметр сортировки не указан, используем сортировку по умолчанию
	if sort == "" {
		sort = "created_at desc"
	}

	sortSplit := strings.Split(sort, "_")

	if len(sortSplit) != 2 {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid sort parameter format")
	}

	orderExpr := getSortExpression(sortSplit[0], sortSplit[1])

	if cursorID == nil {
		rows, err = s.pool.Query(ctx, `
		SELECT `+listingFields+`
        FROM listings l
        INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
        WHERE lsr.title_vector @@ to_tsquery('russian', $1)
        AND l.deleted_at IS NULL
        ORDER BY `+orderExpr+`
		LIMIT $2
		`, createSearchQuery(query), limit)
	} else if limit > 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY `+orderExpr+`) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.title_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT `+listingFields+`
		FROM ranked_listings l
        WHERE row_number >= (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number
		LIMIT $2 + 1
		`, createSearchQuery(query), limit, cursorID)
	} else if limit < 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY `+orderExpr+`) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.title_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT `+listingFields+` 
		FROM ranked_listings l
		WHERE row_number <= (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number DESC
		LIMIT $2 + 1
		`, createSearchQuery(query), -limit, cursorID)
	}

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var listings []models.Listing
	var listingsAnchors *models.Listing

	for rows.Next() {
		var listing models.Listing
		if err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Description,
			&listing.Price,
			&listing.Currency,
			&listing.ViewsCount,
			&listing.CreatedAt,
			&listing.UpdatedAt,
			&listing.DeletedAt,
		); err != nil {
			return nil, nil, err
		}

		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	if limit < 0 {
		slices.Reverse(listings)
	}

	if cursorID != nil {
		listingsAnchors = &listings[0]
		listings = listings[1:]
	}

	return listingsAnchors, listings, nil
}

func (s *Listing) SearchListingsByDescription(ctx context.Context, query string, limit int, cursorID *uuid.UUID, sortOrder string) ([]models.Listing, error) {
	var (
		rows pgx.Rows
		err  error
	)

	sortSplit := strings.Split(sortOrder, "_")

	orderExpr := getSortExpression(sortSplit[0], sortSplit[1])

	if cursorID == nil {
		rows, err = s.pool.Query(ctx, `
		SELECT `+listingFields+`
        FROM listings l
        INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
        WHERE lsr.description_vector @@ to_tsquery('russian', $1)
        AND l.deleted_at IS NULL
        AND NOT EXISTS (
            SELECT 1 FROM listings_search_ru lsr2
            WHERE lsr2.listing_id = l.id
            AND lsr2.title_vector @@ to_tsquery('russian', $1)
        )
        ORDER BY `+orderExpr+`
		LIMIT $2
		`, createSearchQuery(query), limit)
	} else if limit > 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY `+orderExpr+`) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.description_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
            AND NOT EXISTS (
                SELECT 1 FROM listings_search_ru lsr2
                WHERE lsr2.listing_id = l.id
                AND lsr2.title_vector @@ to_tsquery('russian', $1)
            )
		)
		SELECT `+listingFields+`
		FROM ranked_listings l
        WHERE row_number > (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number
		LIMIT $2
		`, createSearchQuery(query), limit, cursorID)
	} else if limit < 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY `+orderExpr+`) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.description_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
            AND NOT EXISTS (
                SELECT 1 FROM listings_search_ru lsr2
                WHERE lsr2.listing_id = l.id
                AND lsr2.title_vector @@ to_tsquery('russian', $1)
            )
		)
		SELECT `+listingFields+`
		FROM ranked_listings l
		WHERE row_number < (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number DESC
		LIMIT $2
		`, createSearchQuery(query), -limit, cursorID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanListings(rows)
}
