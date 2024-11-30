package entity

import (
	"github.com/google/uuid"
	"time"
)

type Poster struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Text  string    `json:"text"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
