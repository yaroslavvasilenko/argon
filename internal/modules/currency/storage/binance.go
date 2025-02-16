package storage

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	client "github.com/binance/binance-connector-go"
	"github.com/yaroslavvasilenko/argon/config"
)

type IBinance interface {
	GetExchangeRateFromBinance(symbol string) (float64, error)
}

type LocalBinance struct{}

func (b *LocalBinance) GetExchangeRateFromBinance(symbol string) (float64, error) {
	// Используем текущее время как seed для рандома
	rand.Seed(time.Now().UnixNano())

	// Генерируем случайное число от 90 до 110
	base := 100.0
	variation := 10.0
	randomPrice := base + (rand.Float64()*variation*2 - variation)

	return randomPrice, nil
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
