package models

import (
	"github.com/google/uuid"
)

type Location struct {
	ID        string    `json:"id" validate:"required,not_blank"`
	ListingID uuid.UUID `json:"-"`
	Name      string    `json:"name" validate:"required,not_blank"`
	Area      Area      `json:"area" validate:"required"`
}

type Area struct {
	Coordinates Coordinates `json:"coordinates" validate:"required"`
	Radius int `json:"radius" validate:"required,min=1"`
}

type Coordinates struct {
	Lat float64 `json:"lat" validate:"required,min=-90,max=90"`
	Lng float64 `json:"lng" validate:"required,min=-180,max=180"`
}
