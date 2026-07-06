// Package idempotency provides a PostgreSQL-backed store for recording
// processed request keys, enabling safe retries without duplicate side-effects.
package idempotency

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store records and checks processed request keys.
type Store interface {
	// Exists returns true if key has already been recorded.
	Exists(ctx context.Context, key string) (bool, error)
	// Record marks key as processed. Idempotent — re-recording the same key is a no-op.
	Record(ctx context.Context, key string) error
}

// PostgresStore persists idempotency keys in a PostgreSQL table.
// The table is created automatically on first call to Init.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a store backed by an existing pgxpool connection pool.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// NewPostgresStoreFromURL opens a new connection pool from connURL and returns a store.
// The caller is responsible for closing the pool when done.
func NewPostgresStoreFromURL(ctx context.Context, connURL string) (*PostgresStore, func(), error) {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		return nil, nil, err
	}
	return &PostgresStore{pool: pool}, pool.Close, nil
}

// Init creates the idempotency_keys table if it does not already exist.
// Call this once at service startup before any Exists/Record calls.
func (s *PostgresStore) Init(ctx context.Context) error {
	const q = `CREATE TABLE IF NOT EXISTS idempotency_keys (
		key        VARCHAR(512) PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
	_, err := s.pool.Exec(ctx, q)
	return err
}

// Exists returns true if key is present in the store.
func (s *PostgresStore) Exists(ctx context.Context, key string) (bool, error) {
	var exists bool
	const q = `SELECT EXISTS(SELECT 1 FROM idempotency_keys WHERE key = $1)`
	err := s.pool.QueryRow(ctx, q, key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Record inserts key into the store. Duplicate inserts are silently ignored.
func (s *PostgresStore) Record(ctx context.Context, key string) error {
	const q = `INSERT INTO idempotency_keys (key) VALUES ($1) ON CONFLICT DO NOTHING`
	_, err := s.pool.Exec(ctx, q, key)
	return err
}
