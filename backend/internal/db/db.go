// Package db provides PostgreSQL connection management.
package db

import (
	"context"
	"fmt"
	"log/slog"
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
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnString()) // ParseConfig validates the connection string and prepares the config, also it uses the URI from .env
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Connection pool tuning
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	slog.Info("connecting to database",
		"host", poolCfg.ConnConfig.Host,
		"port", poolCfg.ConnConfig.Port,
		"database", poolCfg.ConnConfig.Database,
		"sslmode", poolCfg.ConnConfig.TLSConfig != nil,
	)

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	slog.Info("database connection established",
		"max_conns", poolCfg.MaxConns,
		"min_conns", poolCfg.MinConns,
	)

	return &Pool{Pool: pool}, nil
}

// Close closes the connection pool.
func (p *Pool) Close() {
	slog.Info("closing database connection pool")
	p.Pool.Close()
}
