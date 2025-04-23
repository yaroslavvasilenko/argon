package models

import (
	"time"

	"github.com/google/uuid"
)

// ImageLink представляет связь между изображением и объявлением
type ImageLink struct {
	NameImage  string
	ListingID uuid.UUID
	Linked    bool
	UpdatedAt time.Time
}
