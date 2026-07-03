// Package postgres_test contains integration tests for the driver postgres repositories.
// Tests require a running PostgreSQL instance pointed to by DATABASE_URL.
// When DATABASE_URL is not set the entire package is skipped (exit 0).
//
// Start the local database with: make infra-up
// Then run: DATABASE_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride" \
//           go test github.com/fairride/driver/infrastructure/postgres/... -v
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
		CREATE TABLE IF NOT EXISTS driver_profiles (
			driver_id           TEXT        PRIMARY KEY,
			user_id             TEXT        NOT NULL UNIQUE,
			license_number      TEXT        NOT NULL,
			vehicle_type        TEXT        NOT NULL,
			vehicle_brand       TEXT        NOT NULL DEFAULT '',
			vehicle_model       TEXT        NOT NULL DEFAULT '',
			vehicle_color       TEXT        NOT NULL DEFAULT '',
			plate_number        TEXT        NOT NULL,
			online_status       TEXT        NOT NULL DEFAULT 'offline',
			verification_status TEXT        NOT NULL DEFAULT 'pending',
			created_at          TIMESTAMPTZ NOT NULL,
			updated_at          TIMESTAMPTZ NOT NULL
		);
		CREATE TABLE IF NOT EXISTS vehicles (
			vehicle_id   TEXT        PRIMARY KEY,
			driver_id    TEXT        NOT NULL,
			type         TEXT        NOT NULL,
			brand        TEXT        NOT NULL DEFAULT '',
			model        TEXT        NOT NULL DEFAULT '',
			color        TEXT        NOT NULL DEFAULT '',
			plate_number TEXT        NOT NULL,
			year         INT         NOT NULL DEFAULT 0,
			created_at   TIMESTAMPTZ NOT NULL,
			updated_at   TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS vehicles_driver_id_idx ON vehicles(driver_id);
	`)
	return err
}

func dropSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS vehicles;
		DROP TABLE IF EXISTS driver_profiles;
	`)
	return err
}

func setupTest(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE driver_profiles, vehicles`)
	if err != nil {
		t.Fatalf("setupTest: truncate: %v", err)
	}
}
