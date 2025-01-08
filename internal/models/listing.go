package models

import (
	"time"

	"github.com/google/uuid"
)

type Listing struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
	Text  string    `json:"text" gorm:"column:original_description"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
