package currency

import (
	"time"

	"github.com/yaroslavvasilenko/argon/internal/models"
)

type GetCurrencyRequest struct {
	From string `json:"from" validate:"required,oneof=USD EUR RUB ARS"`
	To   string `json:"to" validate:"required,oneof=USD EUR RUB ARS"`
}

type GetCurrencyResponse struct {
	Rate float64 `json:"rate"` // exchange rate
}

type CreateListingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty" gorm:"column:original_description"`

	Price      float64         `json:"price,omitempty"`
	Currency   models.Currency `json:"currency,omitempty"`
	ViewsCount int             `json:"views_count,omitempty"`

	Location models.Location `json:"location,omitempty"`
	Categories []string `json:"categories,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
