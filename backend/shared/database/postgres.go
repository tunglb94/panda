// Package database provides a PostgreSQL connection pool for FAIRRIDE services.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection parameters.
type Config struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// Pool is an alias for pgxpool.Pool, kept as the public type so callers
// can use pgx pool methods without importing pgx directly.
type Pool = pgxpool.Pool

// Connect creates and validates a PostgreSQL connection pool.
// The caller is responsible for closing the pool with pool.Close().
func Connect(ctx context.Context, cfg Config) (*Pool, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

// MustConnect is like Connect but panics on error. Use only during service startup.
func MustConnect(ctx context.Context, cfg Config) *Pool {
	pool, err := Connect(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("fairride/database: %v", err))
	}
	return pool
}
