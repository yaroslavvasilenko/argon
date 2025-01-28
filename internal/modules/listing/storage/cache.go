package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/yaroslavvasilenko/argon/internal/modules/listing"
)

type Cache struct {
	cursors    map[string]CursorInfo
	cursorsMu  sync.RWMutex
	searches   map[string]SearchIdInfo
	searchesMu sync.RWMutex
	secret     []byte
}

func NewCache() *Cache {
	cache := &Cache{
		secret:   []byte("your-secret-key-here"), // в реальном приложении брать из конфига
		cursors:  make(map[string]CursorInfo),
		searches: make(map[string]SearchIdInfo),
	}

	// Запускаем процесс очистки курсоров
	go cache.cleanExpired()

	return cache
}

func (s *Cache) StoreCursor(cursorInfo listing.SearchCursor) string {
	// Сериализуем курсор в JSON
	cursorBytes, err := json.Marshal(cursorInfo)
	if err != nil {
		return ""
	}

	// Создаем HMAC с SHA-256
	h := hmac.New(sha256.New, s.secret)
	h.Write(cursorBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	// Сохраняем информацию о курсоре
	info := CursorInfo{
		Cursor:    cursorInfo,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.cursorsMu.Lock()
	s.cursors[hash] = info
	s.cursorsMu.Unlock()
	return hash
}

func (s *Cache) StoreSearchInfo(searchInfo listing.SearchId) string {
	// Сериализуем в JSON
	searchBytes, err := json.Marshal(searchInfo)
	if err != nil {
		return ""
	}

	// Создаем HMAC с SHA-256
	h := hmac.New(sha256.New, s.secret)
	h.Write(searchBytes)
	hash := hex.EncodeToString(h.Sum(nil))

	// Сохраняем информацию о поиске
	infoSearch := SearchIdInfo{
		SearchId:  searchInfo,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	s.searchesMu.Lock()
	s.searches[hash] = infoSearch
	s.searchesMu.Unlock()
	return hash
}

func (s *Cache) GetCursor(cursorId string) (listing.SearchCursor, error) {
	s.cursorsMu.RLock()
	info, ok := s.cursors[cursorId]
	s.cursorsMu.RUnlock()

	if ok {
		if time.Now().After(info.ExpiresAt) {
			s.cursorsMu.Lock()
			delete(s.cursors, cursorId)
			s.cursorsMu.Unlock()
			return listing.SearchCursor{}, errors.New("cursor expired")
		}
		return info.Cursor, nil
	}
	return listing.SearchCursor{}, errors.New("invalid cursor")
}

func (s *Cache) GetSearchInfo(searchId string) (listing.SearchId, error) {
	s.searchesMu.RLock()
	info, ok := s.searches[searchId]
	s.searchesMu.RUnlock()

	if ok {
		if time.Now().After(info.ExpiresAt) {
			s.searchesMu.Lock()
			delete(s.searches, searchId)
			s.searchesMu.Unlock()
			return listing.SearchId{}, errors.New("search expired")
		}
		return info.SearchId, nil
	}
	return listing.SearchId{}, errors.New("invalid search")
}

func (s *Cache) cleanExpired() {
	ticker := time.NewTicker(time.Hour) // проверяем раз в час
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Очистка устаревших курсоров
		s.cursorsMu.Lock()
		for key, info := range s.cursors {
			if now.After(info.ExpiresAt) {
				delete(s.cursors, key)
			}
		}
		s.cursorsMu.Unlock()

		// Очистка устаревших поисков
		s.searchesMu.Lock()
		for key, info := range s.searches {
			if now.After(info.ExpiresAt) {
				delete(s.searches, key)
			}
		}
		s.searchesMu.Unlock()
	}
}

type CursorInfo struct {
	Cursor    listing.SearchCursor
	ExpiresAt time.Time
}

type SearchIdInfo struct {
	SearchId  listing.SearchId
	ExpiresAt time.Time
}
