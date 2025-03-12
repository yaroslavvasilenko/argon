package storage

import (
	"github.com/jackc/pgx/v5"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

const (
	itemTable     = "listings"
	listingFields = "l.id, l.title, l.original_description, l.price, l.currency, l.views_count, l.created_at, l.updated_at, l.deleted_at"
)

func (s *Listing) scanListings(rows pgx.Rows) ([]models.Listing, error) {
	var listings []models.Listing
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
			return nil, err
		}
		listings = append(listings, listing)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return listings, nil
}
