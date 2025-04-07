package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phuslu/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	gormlog "gorm.io/gorm/logger"
)

type QueryTracer struct {
	log log.Logger
}

func (q *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	q.log.Info().
		Str("query", data.SQL).
		Interface("args", data.Args).
		Msg("SQL query started")
	return ctx
}

func (q *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		q.log.Error().
			Err(data.Err).
			Msg("SQL query failed")
		return
	}
	q.log.Info().
		Int64("rows_affected", data.CommandTag.RowsAffected()).
		Msg("SQL query completed")
}

func NewSqlDB(ctx context.Context, dbUrl string, log log.Logger, debug bool) (*gorm.DB, *pgxpool.Pool, error) {
	gormInstance, err := gorm.Open(
		postgres.Open(dbUrl),
		&gorm.Config{
			Logger: gormlog.Default.LogMode(gormlog.Info),
		})
	if err != nil {
		return nil, nil, err
	}

	confPgx, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, nil, err
	}

	confPgx.ConnConfig.Tracer = &QueryTracer{log: log}

	pool, err := pgxpool.NewWithConfig(ctx, confPgx)
	if err != nil {
		return nil, nil, err
	}

	pool.Config()

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}

	return gormInstance, pool, nil
}
