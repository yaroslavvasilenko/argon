package models

// ListingResult представляет объявление со всеми связанными данными
type ListingResult struct {
	Listing         Listing                      `json:"listing"`
	Categories      []Category                   `json:"categories,omitempty"`
	Boosts          []Boost                      `json:"boosts,omitempty"`
	Characteristics map[string]interface{}       `json:"characteristics,omitempty"`
	Location        Location                     `json:"location,omitempty"`
}

// NewListingResult создает новый экземпляр ListingResult
func NewListingResult(listing Listing) ListingResult {
	return ListingResult{
		Listing:         listing,
		Categories:      []Category{},
		Boosts:          []Boost{},
		Characteristics: map[string]interface{}{},
	}
}

// SetCategories устанавливает категории для объявления
func (r *ListingResult) SetCategories(categories []Category) {
	r.Categories = categories
}

// SetBoosts устанавливает бусты для объявления
func (r *ListingResult) SetBoosts(boosts []Boost) {
	r.Boosts = boosts
}

// SetCharacteristics устанавливает характеристики для объявления
func (r *ListingResult) SetCharacteristics(characteristics map[string]interface{}) {
	r.Characteristics = characteristics
}

// SetLocation устанавливает местоположение для объявления
func (r *ListingResult) SetLocation(location Location) {
	r.Location = location
}
