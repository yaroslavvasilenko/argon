package models

type CategoryNode struct {
	Category Category `json:"category"`
	Subcategories []CategoryNode `json:"subcategories,omitempty"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Image *string `json:"image,omitempty"`
}