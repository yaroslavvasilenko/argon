package entity

import (
	"github.com/google/uuid"
	"time"
)

type Poster struct {
	ID    uuid.UUID
	Title string
	Text  string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
