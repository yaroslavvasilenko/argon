package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"gorm.io/gorm"
)

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

const itemTable = "listings"

func NewStorage(db *gorm.DB, pool *pgxpool.Pool) *Storage {
	return &Storage{gorm: db, pool: pool}
}

func (s *Storage) CreateListing(ctx context.Context, p models.Listing) error {
	err := s.gorm.WithContext(ctx).Create(&p).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetListing(ctx context.Context, pID string) (models.Listing, error) {
	poster := models.Listing{}

	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		First(&poster).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Listing{}, fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return models.Listing{}, err
	}

	return poster, nil
}

func (s *Storage) DeleteListing(ctx context.Context, pID string) error {
	err := s.gorm.Table(itemTable).WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", pID).
		Update("deleted_at", time.Now()).
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateListing(ctx context.Context, p models.Listing) error {
	err := s.gorm.Updates(&p).WithContext(ctx).
		Where("deleted_at IS NULL").
		Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SearchListingsByTitle(ctx context.Context, query string, limit int, cursorID *uuid.UUID) ([]models.Listing, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if cursorID == nil {
		rows, err = s.pool.Query(ctx, `
		SELECT l.*
        FROM listings l
        INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
        WHERE lsr.title_vector @@ to_tsquery('russian', $1)
        AND l.deleted_at IS NULL
        ORDER BY ts_rank(lsr.title_vector, to_tsquery('russian', $1)) DESC
		LIMIT $2
		`, createSearchQuery(query), limit)
	} else if limit > 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY ts_rank(lsr.title_vector, to_tsquery('russian', $1)) DESC) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.title_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT id, title, original_description, created_at, updated_at, deleted_at
		FROM ranked_listings
        WHERE row_number > (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number DESC
		LIMIT $2
		`, createSearchQuery(query), limit, cursorID)
	} else if limit < 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY ts_rank(lsr.title_vector, to_tsquery('russian', $1)) DESC) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.title_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT id, title, original_description, created_at, updated_at, deleted_at
		FROM ranked_listings
		WHERE row_number < (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number ASC
		LIMIT $2
		`, createSearchQuery(query), -limit, cursorID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanListings(rows)
}

func (s *Storage) SearchListingsByDescription(ctx context.Context, query string, limit int, cursorID *uuid.UUID) ([]models.Listing, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if cursorID == nil {
		rows, err = s.pool.Query(ctx, `
		SELECT l.*
        FROM listings l
        INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
        WHERE lsr.description_vector @@ to_tsquery('russian', $1)
        AND l.deleted_at IS NULL
        ORDER BY ts_rank(lsr.description_vector, to_tsquery('russian', $1)) DESC
		LIMIT $2
		`, createSearchQuery(query), limit)
	} else if limit > 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY ts_rank(lsr.description_vector, to_tsquery('russian', $1)) DESC) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.description_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT id, title, original_description, created_at, updated_at, deleted_at
		FROM ranked_listings
        WHERE row_number > (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number DESC
		LIMIT $2
		`, createSearchQuery(query), limit, cursorID)
	} else if limit < 0 {
		rows, err = s.pool.Query(ctx, `
		WITH ranked_listings AS (
			SELECT l.*,
			       ROW_NUMBER() OVER (ORDER BY ts_rank(lsr.description_vector, to_tsquery('russian', $1)) DESC) AS row_number
			FROM listings l
			INNER JOIN listings_search_ru lsr ON l.id = lsr.listing_id
			WHERE lsr.description_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT id, title, original_description, created_at, updated_at, deleted_at
		FROM ranked_listings
		WHERE row_number < (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number ASC
		LIMIT $2
		`, createSearchQuery(query), -limit, cursorID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanListings(rows)
}

func createSearchQuery(query string) string {
    words := strings.Fields(query)
    if len(words) == 0 {
        return ""
    }
    
    if len(words) == 1 {
        // Для одного слова используем префиксный поиск
        return words[0] + ":*"
    }
    
    // Для нескольких слов обрабатываем каждое слово отдельно
    var queryParts []string
    for i, word := range words {
        if i == len(words)-1 {
            // Последнее слово с префиксным поиском
            queryParts = append(queryParts, word + ":*")
        } else {
            // Предыдущие слова ищутся полностью
            queryParts = append(queryParts, word)
        }
    }
    
    // Соединяем слова оператором &
    return strings.Join(queryParts, " & ")
}

func (s *Storage) scanListings(rows pgx.Rows) ([]models.Listing, error) {
	var listings []models.Listing
	for rows.Next() {
		var listing models.Listing
		if err := rows.Scan(
			&listing.ID,
			&listing.Title,
			&listing.Text,
			&listing.CreatedAt,
			&listing.UpdatedAt,
			&listing.DeletedAt,
		); err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return listings, nil
}
