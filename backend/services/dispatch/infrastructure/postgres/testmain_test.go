// Package postgres_test contains integration tests for the dispatch postgres repositories.
// Tests require a running PostgreSQL instance pointed to by DATABASE_URL.
// When DATABASE_URL is not set the entire package is skipped (exit 0).
//
// Start the local database with: make infra-up
// Then run: DATABASE_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride" \
//           go test github.com/fairride/dispatch/infrastructure/postgres/... -v
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

var testNow = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "dispatch postgres tests: DATABASE_URL not set — skipping")
		return 0
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dispatch postgres tests: connect: %v\n", err)
		return 1
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "dispatch postgres tests: ping: %v\n", err)
		return 1
	}

	testPool = pool

	if err := createSchema(ctx, pool); err != nil {
		fmt.Fprintf(os.Stderr, "dispatch postgres tests: createSchema: %v\n", err)
		return 1
	}
	defer func() {
		if err := dropSchema(ctx, pool); err != nil {
			fmt.Fprintf(os.Stderr, "dispatch postgres tests: dropSchema: %v\n", err)
		}
	}()

	return m.Run()
}

func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS trips (
			trip_id             TEXT        PRIMARY KEY,
			rider_id            TEXT        NOT NULL,
			driver_id           TEXT        NOT NULL DEFAULT '',
			status              TEXT        NOT NULL DEFAULT 'pending',
			pickup_address      TEXT        NOT NULL DEFAULT '',
			dropoff_address     TEXT        NOT NULL DEFAULT '',
			cancellation_reason TEXT        NOT NULL DEFAULT '',
			created_at          TIMESTAMPTZ NOT NULL,
			updated_at          TIMESTAMPTZ NOT NULL
		);
		CREATE TABLE IF NOT EXISTS dispatch_jobs (
			job_id             TEXT            PRIMARY KEY,
			trip_id            TEXT            NOT NULL UNIQUE,
			rider_id           TEXT            NOT NULL,
			pickup_lat         DOUBLE PRECISION NOT NULL,
			pickup_lon         DOUBLE PRECISION NOT NULL,
			status             TEXT            NOT NULL DEFAULT 'pending',
			current_driver_id  TEXT            NOT NULL DEFAULT '',
			assigned_driver_id TEXT            NOT NULL DEFAULT '',
			offered_driver_ids TEXT            NOT NULL DEFAULT '',
			offer_expires_at   TIMESTAMPTZ,
			offer_timeout_sec  INT             NOT NULL DEFAULT 30,
			max_attempts       INT             NOT NULL DEFAULT 5,
			attempt_count      INT             NOT NULL DEFAULT 0,
			created_at         TIMESTAMPTZ     NOT NULL,
			updated_at         TIMESTAMPTZ     NOT NULL
		);
		CREATE INDEX IF NOT EXISTS dispatch_jobs_status_expires_idx
			ON dispatch_jobs(offer_expires_at)
			WHERE status = 'searching';
	`)
	return err
}

func dropSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS dispatch_jobs;
		DROP TABLE IF EXISTS trips;
	`)
	return err
}

func setupTest(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE dispatch_jobs, trips`)
	if err != nil {
		t.Fatalf("setupTest: truncate: %v", err)
	}
}
