package models

import "github.com/google/uuid"

type Seller struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Rating     int       `json:"rating,omitempty"`
	Votes      int       `json:"votes,omitempty"`
	Avalilable bool      `json:"avalilable,omitempty"`
	Image      string    `json:"image,omitempty"`
}
