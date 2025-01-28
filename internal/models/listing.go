package models

import (
	"time"

	"github.com/google/uuid"
)

type Currency string

const (
	USD Currency = "USD"
	RUB Currency = "RUB"
	ARS Currency = "ARS"
	EUR Currency = "EUR"
)

type Listing struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description" gorm:"column:original_description"`

	Price      float64  `json:"price"`
	Currency   Currency `json:"currency"`
	ViewsCount int      `json:"views_count"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
