package controller

import (
	"github.com/yaroslavvasilenko/argon/internal/modules/currency"
	"github.com/yaroslavvasilenko/argon/internal/modules/currency/service"
	"github.com/gofiber/fiber/v2"
)

type Currency struct {
	s *service.Currency
}

func NewCurrency(s *service.Currency) *Currency {
	return &Currency{s: s}
}


func (с *Currency) GetCurrency(c *fiber.Ctx) error {
	req := currency.GetCurrencyRequest{}
	if err := c.QueryParser(&req); err != nil {
		return err
	}

	currency, err := с.s.GetCurrency(c.UserContext(), req)
	if err != nil {
		return err
	}
	
	return c.JSON(currency)
}