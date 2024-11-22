package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gormlog "gorm.io/gorm/logger"
)

func NewSqlDB(ctx context.Context, dbUrl string) (*gorm.DB, *pgxpool.Pool, error) {
	gormInstance, err := gorm.Open(
		postgres.Open(dbUrl),
		&gorm.Config{
			Logger: gormlog.Default.LogMode(gormlog.Info),
		})
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, nil, err
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		return nil, nil, err
	}

	return gormInstance, pool, nil
}
