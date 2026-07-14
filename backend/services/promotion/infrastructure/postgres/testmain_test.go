// Package postgres_test contains integration tests for the promotion
// postgres repository. Tests require a running PostgreSQL instance pointed
// to by DATABASE_URL. When DATABASE_URL is not set the entire package is
// skipped (exit 0).
//
// Start the local database with: make infra-up
// Then run: DATABASE_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride" \
//           go test github.com/fairride/promotion/infrastructure/postgres/... -v
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
		CREATE TABLE IF NOT EXISTS vouchers (
			id                 TEXT        PRIMARY KEY,
			code               TEXT        NOT NULL DEFAULT '',
			name               TEXT        NOT NULL,
			description        TEXT        NOT NULL DEFAULT '',
			status             TEXT        NOT NULL,
			priority           INT         NOT NULL DEFAULT 0,
			start_time         TIMESTAMPTZ NOT NULL,
			end_time           TIMESTAMPTZ NOT NULL,
			max_usage          BIGINT      NOT NULL DEFAULT 0,
			max_usage_per_user BIGINT      NOT NULL DEFAULT 0,
			budget             BIGINT      NOT NULL,
			remaining_budget   BIGINT      NOT NULL,
			discount_type      TEXT        NOT NULL,
			discount_value     BIGINT      NOT NULL,
			max_discount       BIGINT      NOT NULL DEFAULT 0,
			min_order          BIGINT      NOT NULL DEFAULT 0,
			vehicle_types      TEXT[]      NOT NULL DEFAULT '{}',
			cities             TEXT[]      NOT NULL DEFAULT '{}',
			membership         TEXT[]      NOT NULL DEFAULT '{}',
			service_types      TEXT[]      NOT NULL DEFAULT '{}',
			trip_types         TEXT[]      NOT NULL DEFAULT '{}',
			campaign           TEXT        NOT NULL DEFAULT '',
			new_user_only      BOOLEAN     NOT NULL DEFAULT FALSE,
			combinable         BOOLEAN     NOT NULL DEFAULT FALSE,
			stackable          BOOLEAN     NOT NULL DEFAULT FALSE,
			promotion_type     TEXT        NOT NULL,
			usage_count        BIGINT      NOT NULL DEFAULT 0,
			created_at         TIMESTAMPTZ NOT NULL,
			updated_at         TIMESTAMPTZ NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS vouchers_code_idx ON vouchers (lower(code)) WHERE code <> '';

		CREATE TABLE IF NOT EXISTS voucher_redemptions (
			id              BIGSERIAL   PRIMARY KEY,
			voucher_id      TEXT        NOT NULL REFERENCES vouchers(id),
			rider_id        TEXT        NOT NULL,
			trip_id         TEXT        NOT NULL,
			discount_amount BIGINT      NOT NULL,
			status          TEXT        NOT NULL DEFAULT 'redeemed',
			created_at      TIMESTAMPTZ NOT NULL,
			updated_at      TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS voucher_redemptions_voucher_rider_idx ON voucher_redemptions(voucher_id, rider_id);
	`)
	return err
}

func dropSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS voucher_redemptions;
		DROP TABLE IF EXISTS vouchers;
	`)
	return err
}

func setupTest(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE voucher_redemptions, vouchers`)
	if err != nil {
		t.Fatalf("setupTest: truncate: %v", err)
	}
}
