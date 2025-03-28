package storage

import (
	"context"
	"fmt"
	"strconv"

	client "github.com/binance/binance-connector-go"
	"github.com/yaroslavvasilenko/argon/config"
)

type IBinance interface {
	GetExchangeRateFromBinance(symbol string) (float64, error)
}

type LocalBinance struct{}

func (b *LocalBinance) GetExchangeRateFromBinance(symbol string) (float64, error) {
	// Захардкоженные курсы валют (примерно реальные на март 2025)
	exchangeRates := map[string]float64{
		// Основные пары с USD
		"EURUSD": 1.08,  // EUR/USD
		"USDRUB": 92.5,  // USD/RUB
		"USDARS": 950.0, // USD/ARS

		// Обратные пары
		"USDEUR": 0.925,   // USD/EUR
		"RUBUSD": 0.0108,  // RUB/USD
		"ARSUSD": 0.00105, // ARS/USD

		// Кросс-курсы EUR
		"EURRUB": 100.0,   // EUR/RUB
		"EURARS": 1027.0,  // EUR/ARS
		"RUBEUR": 0.01,    // RUB/EUR
		"ARSEUR": 0.00097, // ARS/EUR

		// Кросс-курсы RUB и ARS
		"RUBARS": 10.27,  // RUB/ARS
		"ARSRUB": 0.0974, // ARS/RUB
	}

	// Проверяем, есть ли запрошенный символ в нашей таблице курсов
	rate, exists := exchangeRates[symbol]
	if !exists {
		return 0, fmt.Errorf("курс для пары %s не найден", symbol)
	}

	return rate, nil
}

type Binance struct {
	client *client.Client
}

func NewBinance(cfg config.Config) IBinance {
	if cfg.Binance.Local {
		return &LocalBinance{}
	}
	if cfg.Binance.APIKey == "" || cfg.Binance.SecretKey == "" {
		panic("необходимо указать API ключи для Binance")
	}

	c := client.NewClient(cfg.Binance.APIKey, cfg.Binance.SecretKey, "")
	return &Binance{
		client: c,
	}
}

func (c *Binance) GetExchangeRateFromBinance(symbol string) (float64, error) {
	// Получаем текущую цену используя клиент Binance
	resp, err := c.client.NewTickerPriceService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, fmt.Errorf("ошибка запроса к Binance: %v", err)
	}

	if len(resp) == 0 {
		return 0, fmt.Errorf("нет данных о цене")
	}

	price, err := strconv.ParseFloat(resp[0].Price, 64)
	if err != nil {
		return 0, fmt.Errorf("ошибка конвертации цены: %v", err)
	}

	return price, nil
}
