// Package db provides PostgreSQL connection management.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool for database operations.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new connection pool to PostgreSQL.
func NewPool(ctx context.Context, cfg config.DBConfig) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	poolCfg.MaxConns = 20
	poolCfg.MinConns = 5
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Close closes the connection pool.
func (p *Pool) Close() {
	p.Pool.Close()
}
