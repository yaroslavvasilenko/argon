package storage

import (
	"context"
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

type Storage struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

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

func (s *Storage) SearchListingsByTitle(ctx context.Context, query string, limit int, cursorID *uuid.UUID, sort string) ([]models.Listing, error) {
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
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid sort parameter format")
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
			WHERE lsr.title_vector @@ to_tsquery('russian', $1)
			AND l.deleted_at IS NULL
		)
		SELECT id, title, original_description as description, price, currency, views_count, created_at, updated_at, deleted_at
		FROM ranked_listings
		WHERE row_number < (SELECT row_number FROM ranked_listings WHERE id = $3)
		ORDER BY row_number DESC
		LIMIT $2
		`, createSearchQuery(query), -limit, cursorID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res, err := s.scanListings(rows)
	if err != nil {
		return nil, err
	}

	if limit < 0 {
		slices.Reverse(res)
	}

	return res, nil
}

func (s *Storage) SearchListingsByDescription(ctx context.Context, query string, limit int, cursorID *uuid.UUID, sortOrder string) ([]models.Listing, error) {
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
		FROM ranked_listings
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
		FROM ranked_listings
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
