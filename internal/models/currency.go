package models

import (
	"fmt"
	"time"
)

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

// IsValid проверяет, является ли валюта допустимой
func (c Currency) IsValid() bool {
	for _, validCurrency := range Currencies {
		if c == validCurrency {
			return true
		}
	}
	return false
}

// Validate проверяет валидность валюты и возвращает ошибку, если валюта недопустима
func (c Currency) Validate() error {
	if c == "" {
		return fmt.Errorf("currency is required")
	}
	
	if !c.IsValid() {
		return fmt.Errorf("invalid currency: %s. Supported currencies: %v", c, Currencies)
	}
	
	return nil
}


type ExchangeRate struct {
	Symbol        Currency
	QuoteSymbol   Currency
	ExchangeRate  float64 
	ExpiresAt     time.Time  
	CreatedAt     time.Time  
	UpdatedAt     *time.Time  
}