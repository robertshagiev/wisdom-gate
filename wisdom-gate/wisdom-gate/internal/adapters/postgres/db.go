package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDBPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	const op = "adapters.postgres.NewPostgresDBPool"

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot parse DB config: %w", op, err)
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol // disable prepared queries for pgbouncer
	poolCfg.MaxConnIdleTime = time.Minute * 5
	poolCfg.MaxConnLifetime = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot connect to DB: %w", op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %w", op, err)
	}

	return pool, nil
}
