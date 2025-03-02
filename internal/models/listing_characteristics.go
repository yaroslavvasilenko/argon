package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// ListingCharacteristics представляет характеристики объявления в базе данных
type ListingCharacteristics struct {
	ListingID      uuid.UUID       `json:"listing_id"`
	Characteristics json.RawMessage `json:"characteristics"`
}

// Метод для получения значения характеристики по ключу
func (lc *ListingCharacteristics) GetCharacteristic(key string) (interface{}, bool) {
	var characteristics map[string]interface{}
	if err := json.Unmarshal(lc.Characteristics, &characteristics); err != nil {
		return nil, false
	}
	
	value, exists := characteristics[key]
	return value, exists
}

// Метод для установки значения характеристики
func (lc *ListingCharacteristics) SetCharacteristic(key string, value interface{}) error {
	var characteristics map[string]interface{}
	
	// Если характеристики уже существуют, распаковываем их
	if lc.Characteristics != nil && len(lc.Characteristics) > 0 {
		if err := json.Unmarshal(lc.Characteristics, &characteristics); err != nil {
			return err
		}
	} else {
		// Иначе создаем новую пустую карту
		characteristics = make(map[string]interface{})
	}
	
	// Устанавливаем или обновляем значение
	characteristics[key] = value
	
	// Упаковываем обратно в JSON
	data, err := json.Marshal(characteristics)
	if err != nil {
		return err
	}
	
	lc.Characteristics = data
	return nil
}
