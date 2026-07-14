// Package postgres_test contains integration tests for the postgres repositories.
// Tests require a running PostgreSQL instance pointed to by DATABASE_URL.
// When DATABASE_URL is not set the entire package is skipped (exit 0).
//
// Start the local database with: make infra-up
// Then run: DATABASE_URL="postgres://fairride:fairride_dev_secret@localhost:5432/fairride" \
//           go test github.com/fairride/identity/infrastructure/postgres/... -v
package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// testPool is shared across all tests in this package.
// It is initialised in TestMain and must not be closed before m.Run() returns.
var testPool *pgxpool.Pool

// testNow is a stable timestamp used to construct entities in tests.
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

// createSchema creates the tables required by the Identity persistence layer.
// This is test scaffolding only — NOT a migration. Production table creation
// is deferred until the migration framework is in place (Phase 2+).
func createSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS identity_permissions (
			id          TEXT        PRIMARY KEY,
			name        TEXT        NOT NULL UNIQUE,
			resource    TEXT        NOT NULL,
			action      TEXT        NOT NULL,
			description TEXT        NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS identity_roles (
			id          TEXT        PRIMARY KEY,
			name        TEXT        NOT NULL UNIQUE,
			description TEXT        NOT NULL DEFAULT '',
			is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at  TIMESTAMPTZ NOT NULL,
			updated_at  TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS identity_role_permissions (
			role_id       TEXT NOT NULL REFERENCES identity_roles(id)       ON DELETE CASCADE,
			permission_id TEXT NOT NULL REFERENCES identity_permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		);

		CREATE TABLE IF NOT EXISTS identity_users (
			id             TEXT        PRIMARY KEY,
			phone_number   TEXT        NOT NULL DEFAULT '',
			name           TEXT        NOT NULL,
			email          TEXT        NOT NULL DEFAULT '',
			google_sub     TEXT        NOT NULL DEFAULT '',
			type           TEXT        NOT NULL,
			status         TEXT        NOT NULL,
			role_id        TEXT        NOT NULL,
			driver_enabled BOOLEAN     NOT NULL DEFAULT FALSE,
			created_at     TIMESTAMPTZ NOT NULL,
			updated_at     TIMESTAMPTZ NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_phone_unique
			ON identity_users (phone_number) WHERE phone_number <> '';
		CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_email_unique
			ON identity_users (email) WHERE email <> '';
		CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_google_sub_unique
			ON identity_users (google_sub) WHERE google_sub <> '';
	`)
	return err
}

// dropSchema removes the test tables in dependency order.
func dropSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS identity_users;
		DROP TABLE IF EXISTS identity_role_permissions;
		DROP TABLE IF EXISTS identity_roles;
		DROP TABLE IF EXISTS identity_permissions;
	`)
	return err
}

// setupTest truncates all Identity tables so each test begins with a clean slate.
// Call this at the start of every test that writes to the database.
func setupTest(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		`TRUNCATE identity_users, identity_role_permissions, identity_roles, identity_permissions`)
	if err != nil {
		t.Fatalf("setupTest: truncate: %v", err)
	}
}
