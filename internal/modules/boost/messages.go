package boost

import (
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/models"
)

type GetBoostRequest struct {
	ListingID uuid.UUID `json:"listing_id"`
}

type GetBoostResponse struct {
	AvailableBoosts []BoostResp `json:"boosts"`
	EnableBoostTypes []models.BoostType `json:"enable_boost_types,omitempty"`
}

type BoostResp struct {
	Type              models.BoostType `json:"type"`
	CommissionPercent float64          `json:"commission_percents"`
}

type UpdateBoostRequest struct {
	ListingID uuid.UUID `json:"listing_id"`
	Boosts    []models.BoostType `json:"enabled_boost_types"`
}
