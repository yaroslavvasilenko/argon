package service

import (

	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/currency/storage"
	"github.com/yaroslavvasilenko/argon/internal/modules/currency"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"context"

)

type Currency struct {
	s      *storage.Currency
	logger *logger.LogPhuslu
	b      storage.IBinance
}

func NewCurrency(s *storage.Currency, b storage.IBinance, logger *logger.LogPhuslu) *Currency {
	srv := &Currency{
		s:      s,
		logger: logger,
		b:      b,
	}

	go srv.runHourlySync()

	return srv
}

func (c *Currency) GetCurrency(ctx context.Context, req currency.GetCurrencyRequest) (*currency.GetCurrencyResponse, error) {
	from := models.CurrencyMap[req.From]
	to := models.CurrencyMap[req.To]

	if from == "" || to == "" {
		return nil, nil
	}

	currencyInst, err := c.s.GetCurrency(ctx, from + to)
	if err != nil {
		return nil, err
	}

	return &currency.GetCurrencyResponse{
		Rate: currencyInst.ExchangeRate,
	}, nil
}

