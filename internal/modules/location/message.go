package location

type GetLocationRequest struct {
	Area     Area   `json:"area" validate:"required"`
	Language string `json:"language,omitempty"`
}

type Area struct{
	Coordinates struct{
		Lat float64 `json:"lat" validate:"required"`
		Lng float64 `json:"lng" validate:"required"`
	} `json:"coordinates" validate:"required"`
	Radius int `json:"radius" validate:"required"`
}


type GetLocationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Area Area `json:"area"`
}