package models

import (
	"time"

	"github.com/google/uuid"
)
// Listing представляет объявление в базе данных
type Listing struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty" gorm:"column:original_description"`

	Price      float64   `json:"price,omitempty"`
	Currency   Currency  `json:"currency,omitempty"`
	ViewsCount int       `json:"views_count,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
