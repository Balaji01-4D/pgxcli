package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connects to the PostgreSQL database using the provided DSN (Data Source Name)
// ex: postgres://user:password@localhost:5432/mydb
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxConnIdleTime == 0 {
   		cfg.MaxConnIdleTime = 5 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return pool, nil

}