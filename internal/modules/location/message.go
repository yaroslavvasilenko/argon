package location

import (
	"github.com/yaroslavvasilenko/argon/internal/models"
)

type GetLocationRequest struct {
	Area     models.Area   `json:"area" validate:"required"`
}




type GetLocationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Area models.Area `json:"area"`
}