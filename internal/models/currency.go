package models

import "time"

type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	RUB Currency = "RUB"
	ARS Currency = "ARS"
)

var CurrencyMap = map[string]Currency{
	"USD": USD,
	"EUR": EUR,
	"RUB": RUB,
	"ARS": ARS,
}


var Currencies = []Currency{USD, EUR, RUB, ARS}


type ExchangeRate struct {
	Symbol        Currency
	QuoteSymbol   Currency
	ExchangeRate  float64 
	ExpiresAt     time.Time  
	CreatedAt     time.Time  
	UpdatedAt     *time.Time  
}