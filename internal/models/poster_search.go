package models

import "github.com/google/uuid"

type ListingSearch struct {
	ID    uuid.UUID `json:"-"`
	Title string    `json:"title"`
}

func NewListingSearch(listings []Listing) []ListingSearch {
	listingSearch := make([]ListingSearch, 0, len(listings))
	for _, p := range listings {
		listingSearch = append(listingSearch, ListingSearch{
			ID:    p.ID,
			Title: p.Title,
		})
	}

	return listingSearch

}
