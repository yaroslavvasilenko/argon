package models

import (
	"github.com/google/uuid"
)

type Location struct {
	ID        uuid.UUID `json:"-"`
	ListingID uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	Area      Area      `json:"area"`
}

type Area struct {
	Coordinates struct {
		Lat float64 `json:"lat" validate:"required"`
		Lng float64 `json:"lng" validate:"required"`
	} `json:"coordinates" validate:"required"`
	Radius int `json:"radius" validate:"required"`
}
