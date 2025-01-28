package storage

import (
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

const (
	itemTable     = "listings"
	listingFields = "l.id, l.title, l.original_description, l.price, l.currency, l.views_count, l.created_at, l.updated_at, l.deleted_at"
)

func (s *Storage) scanListings(rows pgx.Rows) ([]models.Listing, error) {
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

func getSortExpression(sortBy string, sortOrder string) string {
	var orderExpr string
	switch sortBy {
	case "price":
		orderExpr = "l.price"
	case "popularity":
		orderExpr = "l.views_count"
	default: // relevance or unknown
		orderExpr = "ts_rank(lsr.title_vector, to_tsquery('russian', $1))"
	}

	if strings.EqualFold(sortOrder, "asc") {
		return orderExpr + " ASC"
	}
	return orderExpr + " DESC"
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
			queryParts = append(queryParts, word+":*")
		} else {
			// Предыдущие слова ищутся полностью
			queryParts = append(queryParts, word)
		}
	}

	// Соединяем слова оператором &
	return strings.Join(queryParts, " & ")
}
