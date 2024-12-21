package models

import "github.com/google/uuid"

type ItemSearch struct {
	ID    uuid.UUID `json:"-"`
	Title string    `json:"title"`
}

func NewItemSearch(posters []Item) []ItemSearch {
	itemSearch := make([]ItemSearch, 0, len(posters))
	for _, p := range posters {
		itemSearch = append(itemSearch, ItemSearch{
			ID:    p.ID,
			Title: p.Title,
		})
	}

	return itemSearch

}
