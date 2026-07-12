-- Shared: backend/shared/idempotency/store.go PostgresStore.Init() — used by
-- the booking service (backend/services/booking/cmd/server/main.go) to record
-- processed request keys. Already a production schema-creation function,
-- reproduced here verbatim so a fresh database has it before any service starts.

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key        VARCHAR(512) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
