package internal

import "gorm.io/gorm"

type Storage struct {
	gorm *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{gorm: db}
}
