package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

type Cache struct {
	pool   *pgxpool.Pool
	secret []byte
}

func NewCache(pool *pgxpool.Pool) *Cache {

	cache := &Cache{
		pool:   pool,
		secret: []byte("your-secret-key-here"), // в реальном приложении брать из конфига
	}

	// Запускаем процесс очистки устаревших записей
	go cache.cleanExpired()

	return cache
}

func (s *Cache) StoreCursor(cursorInfo listing.SearchCursor) string {
	cursorBytes, err := json.Marshal(cursorInfo)
	if err != nil {
		return ""
	}

	// Создаем HMAC с SHA-256
	h := hmac.New(sha256.New, s.secret)
	h.Write(cursorBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	// Сохраняем в базу
	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO search_cursors (id, cursor_data, expires_at) VALUES ($1, $2, $3) "+
			"ON CONFLICT (id) DO UPDATE SET cursor_data = $2, expires_at = $3",
		hash,
		cursorBytes,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return ""
	}

	return hash
}



func (s *Cache) GetCursor(cursorId string) (listing.SearchCursor, error) {
	var cursorBytes []byte
	var expiresAt time.Time

	err := s.pool.QueryRow(context.Background(),
		"SELECT cursor_data, expires_at FROM search_cursors WHERE id = $1",
		cursorId,
	).Scan(&cursorBytes, &expiresAt)

	if err != nil {
		return listing.SearchCursor{}, errors.New("invalid cursor")
	}

	if time.Now().After(expiresAt) {
		// Удаляем устаревший курсор
		_, _ = s.pool.Exec(context.Background(),
			"DELETE FROM search_cursors WHERE id = $1",
			cursorId,
		)
		return listing.SearchCursor{}, errors.New("cursor expired")
	}

	var cursor listing.SearchCursor
	if err := json.Unmarshal(cursorBytes, &cursor); err != nil {
		return listing.SearchCursor{}, errors.New("invalid cursor data")
	}

	return cursor, nil
}

func (s *Cache) StoreSearchInfo(searchInfo listing.SearchID) string {
	searchBytes, err := json.Marshal(searchInfo)
	if err != nil {
		return ""
	}

	// Создаем HMAC с SHA-256
	h := hmac.New(sha256.New, s.secret)
	h.Write(searchBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	// Сохраняем в базу
	_, err = s.pool.Exec(context.Background(),
		"INSERT INTO search_info (id, search_data, expires_at) VALUES ($1, $2, $3) "+
			"ON CONFLICT (id) DO UPDATE SET search_data = $2, expires_at = $3",
		hash,
		searchBytes,
		time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return ""
	}

	return hash
}

func (s *Cache) GetSearchInfo(searchId string) (*listing.SearchID, error) {
	var searchBytes []byte
	var expiresAt time.Time

	err := s.pool.QueryRow(context.Background(),
		"SELECT search_data, expires_at FROM search_info WHERE id = $1",
		searchId,
	).Scan(&searchBytes, &expiresAt)

	if err != nil {
		return nil, errors.New("invalid search id")
	}

	if time.Now().After(expiresAt) {
		// Удаляем устаревшую информацию о поиске
		_, _ = s.pool.Exec(context.Background(),
			"DELETE FROM search_info WHERE id = $1",
			searchId,
		)
		return nil, nil
	}

	searchInfo := &listing.SearchID{}
	if err := json.Unmarshal(searchBytes, searchInfo); err != nil {
		return nil, errors.New("invalid search info data")
	}

	return searchInfo, nil
}

func (s *Cache) cleanExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		_, _ = s.pool.Exec(context.Background(),
			"DELETE FROM search_cursors WHERE expires_at < $1",
			time.Now(),
		)
		_, _ = s.pool.Exec(context.Background(),
			"DELETE FROM search_info WHERE expires_at < $1",
			time.Now(),
		)
	}
}
