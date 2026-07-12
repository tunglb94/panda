-- Service: driver
-- Reconstructed from backend/services/driver/infrastructure/postgres/testmain_test.go
-- (createSchema), cross-checked against driver_repository.go (driverFields) and
-- vehicle_repository.go (vehicleFields) — both match exactly.

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
