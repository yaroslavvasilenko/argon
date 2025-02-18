package service

import (
	"context"
	"time"

	"github.com/yaroslavvasilenko/argon/internal/models"
)

func (c *Currency) runHourlySync() {
	// Запускаем сразу при старте
	if err := c.SyncBinance(); err != nil {
		c.logger.Errorf("ошибка синхронизации курса")
	}

	// Запускаем таймер на каждый час
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.SyncBinance(); err != nil {
				c.logger.Errorf("ошибка синхронизации курса")
			}
		}
	}
}

func (c *Currency) SyncBinance() error {
	// Создаем все возможные пары валют
	for _, base := range models.Currencies {
		for _, quote := range models.Currencies {
			// Пропускаем одинаковые валюты
			if base == quote {
				continue
			}

			// Формируем символ пары для Binance (например USDEUR)
			symbolPair := string(base) + string(quote)

			// Добавляем задержку между запросами (1200 запросов в минуту максимум для Binance)
			time.Sleep(110 * time.Millisecond)

			rate, err := c.b.GetExchangeRateFromBinance(symbolPair)
			if err != nil {
				c.logger.Errorf("ошибка получения курса для пары")
				// При ошибке делаем более длительную паузу
				time.Sleep(500 * time.Millisecond)
				continue // Продолжаем со следующей парой при ошибке
			}

			err = c.s.CreateOrUpdateCurrency(context.Background(), models.ExchangeRate{
				Symbol:       models.Currency(symbolPair),
				QuoteSymbol:  quote,
				ExchangeRate: rate,
			})
			if err != nil {
				c.logger.Errorf("ошибка сохранения курса для пары")
				continue
			}
		}
	}

	return nil
}
