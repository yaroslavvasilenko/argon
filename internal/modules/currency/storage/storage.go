package storage

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"context"
	"time"
	"gorm.io/gorm"
)

const (
	currencyTable = "currency_exchange_rates"
)

type Currency struct {
	gorm *gorm.DB
	pool *pgxpool.Pool
}

func NewCurrency(db *gorm.DB, pool *pgxpool.Pool) *Currency {
	return &Currency{gorm: db, pool: pool}
}



func (s *Currency) CreateOrUpdateCurrency(ctx context.Context, p models.ExchangeRate) error {
	timeNow := time.Now()
	p.CreatedAt = timeNow
	p.UpdatedAt = &timeNow
	p.ExpiresAt = timeNow.Add(time.Hour * 1)

	query := `INSERT INTO currency_exchange_rates 
	(symbol, exchange_rate, expires_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (symbol) DO UPDATE 
	SET exchange_rate = EXCLUDED.exchange_rate, expires_at = EXCLUDED.expires_at, updated_at = EXCLUDED.updated_at`
	_, err := s.pool.Exec(ctx, query, p.Symbol, p.ExchangeRate, p.ExpiresAt, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *Currency) GetCurrency(ctx context.Context, pID models.Currency) (*models.ExchangeRate, error) {
    query := `SELECT symbol, exchange_rate, expires_at, created_at, updated_at FROM currency_exchange_rates WHERE symbol = $1`
    exchangeRate := &models.ExchangeRate{}
    err := s.pool.QueryRow(ctx, query, pID).Scan(
        &exchangeRate.Symbol,
        &exchangeRate.ExchangeRate,
        &exchangeRate.ExpiresAt,
        &exchangeRate.CreatedAt,
        &exchangeRate.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return exchangeRate, nil
}