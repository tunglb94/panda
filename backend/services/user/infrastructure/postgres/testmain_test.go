// Package postgres_test contains integration tests for the user postgres repositories.
// Tests require a running PostgreSQL instance pointed to by DATABASE_URL.
// When DATABASE_URL is not set the entire package is skipped (exit 0).
//
// Start the local database with: make infra-up
// Then run: DATABASE_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride" \
//           go test github.com/fairride/user/infrastructure/postgres/... -v
package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

var testNow = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "postgres tests: DATABASE_URL not set — skipping")
		return 0
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres tests: connect: %v\n", err)
		return 1
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "postgres tests: ping: %v\n", err)
		return 1
	}

	testPool = pool

	if err := createSchema(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "postgres tests: createSchema: %v\n", err)
		return 1
	}
	defer func() {
		if err := dropSchema(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "postgres tests: dropSchema: %v\n", err)
		}
	}()

	return m.Run()
}

func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS user_profiles (
			id           TEXT        PRIMARY KEY,
			full_name    TEXT        NOT NULL,
			phone        TEXT        NOT NULL,
			email        TEXT        NOT NULL DEFAULT '',
			avatar       TEXT        NOT NULL DEFAULT '',
			date_of_birth TIMESTAMPTZ NULL,
			gender       TEXT        NOT NULL DEFAULT 'unspecified',
			status       TEXT        NOT NULL DEFAULT 'active',
			created_at   TIMESTAMPTZ NOT NULL,
			updated_at   TIMESTAMPTZ NOT NULL
		);
	`)
	return err
}

func dropSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `DROP TABLE IF EXISTS user_profiles;`)
	return err
}

func setupTest(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE user_profiles`)
	if err != nil {
		t.Fatalf("setupTest: truncate: %v", err)
	}
}
