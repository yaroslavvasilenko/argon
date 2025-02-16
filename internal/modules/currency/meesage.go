package currency


type GetCurrencyRequest struct {
	From string `json:"from" validate:"required,oneof=USD EUR RUB ARS"`
	To   string `json:"to" validate:"required,oneof=USD EUR RUB ARS"`
}

type GetCurrencyResponse struct {
	Rate float64 `json:"rate"` // exchange rate
}