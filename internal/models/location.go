package models

import (
	"github.com/google/uuid"
)

type Location struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	ListingID uuid.UUID `json:"listing_id" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	Area      Area      `json:"area" validate:"required"`
}

type Area struct {
	Coordinates struct {
		Lat float64 `json:"lat" validate:"required"`
		Lng float64 `json:"lng" validate:"required"`
	} `json:"coordinates" validate:"required"`
	Radius int `json:"radius" validate:"required"`
}