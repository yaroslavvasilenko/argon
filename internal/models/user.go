package models

// Пакет models содержит модели данных для приложения

// User представляет модель пользователя в системе
type User struct {
	ID                 string        `json:"id" validate:"required"`
	Name               string        `json:"name" validate:"required,max=50"`
	DescriptionMap     map[string]string `json:"description,omitempty" validate:"omitempty,max=5000"`
	OriginalDescription *string       `json:"original_description,omitempty" validate:"omitempty,max=5000"`
	LocationID         string        `json:"location_id" validate:"required"`
	Rating             float64      `json:"rating,omitempty" validate:"omitempty,min=0,max=5"`
	Votes              int          `json:"votes,omitempty"`
	Available          bool         `json:"available,omitempty"`
	Images             []string      `json:"images" validate:"required"`
	Editable           bool         `json:"editable,omitempty"`

	ZitadelID string
}

// UserContactType представляет тип контакта пользователя
type UserContactType string

// Константы для типов контактов пользователя
const (
	UserContactTypePhone  UserContactType = "phone"  // Телефон
	UserContactTypeEmail  UserContactType = "email"  // Email
	UserContactTypeTg     UserContactType = "tg"     // Telegram
	UserContactTypeWa     UserContactType = "wa"     // WhatsApp
	UserContactTypeIg     UserContactType = "ig"     // Instagram
	UserContactTypeFb     UserContactType = "fb"     // Facebook
	UserContactTypeMsg    UserContactType = "msg"    // Сообщения на сайте
	UserContactTypeOther  UserContactType = "other"  // Другое
)

// UserContact представляет контактную информацию пользователя
type UserContact struct {
	ID     string          `json:"id" validate:"required"`
	UserID string          `json:"user_id" gorm:"column:user_id"`
	Type   UserContactType `json:"type" validate:"required"`
	Text   string          `json:"text" validate:"required"`
	Link   string          `json:"link,omitempty"`
}
