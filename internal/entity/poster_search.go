package entity

import "github.com/google/uuid"

type PosterSearch struct {
	ID    uuid.UUID `json:"-"`
	Title string    `json:"title"`
}

func NewPosterSearch(posters []Poster) []PosterSearch {
	posterSearch := make([]PosterSearch, 0, len(posters))
	for _, p := range posters {
		posterSearch = append(posterSearch, PosterSearch{
			ID:    p.ID,
			Title: p.Title,
		})
	}

	return posterSearch

}
