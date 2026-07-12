-- Service: trip (table also written to by dispatch's TripUpdater — see 005_dispatch.up.sql)
--
-- Base shape reconstructed from backend/services/trip/infrastructure/postgres/testmain_test.go
-- (createSchema) and the trips_rider_id_idx index defined there.
--
-- MISMATCH DISCOVERED (see sprint report): trip_repository.go's Save/FindByID/
-- FindByRiderID/FindByDriverID all read and write a `payment_method` column
-- that testmain_test.go's own createSchema does NOT define. Since the test
-- schema is stale relative to the repository it's meant to support (and the
-- repository code is what actually runs in production), `payment_method` is
-- included here as the source of truth, following the same NOT NULL DEFAULT ''
-- pattern already used for the other optional string columns in this table
-- (cancellation_reason, fare_currency).
--
-- dispatch/infrastructure/postgres/trip_updater.go only ever UPDATEs
-- `status`, `driver_id`, `updated_at` on this table (never SELECTs, never
-- INSERTs) — fully satisfied by this shape, so no dispatch-specific trips
-- columns are needed here.

CREATE TABLE IF NOT EXISTS trips (
    trip_id             TEXT        PRIMARY KEY,
    rider_id            TEXT        NOT NULL,
    driver_id           TEXT        NOT NULL DEFAULT '',
    status              TEXT        NOT NULL DEFAULT 'pending',
    pickup_address      TEXT        NOT NULL,
    dropoff_address     TEXT        NOT NULL,
    cancellation_reason TEXT        NOT NULL DEFAULT '',
    final_fare_total    BIGINT      NOT NULL DEFAULT 0,
    fare_currency       TEXT        NOT NULL DEFAULT '',
    payment_method      TEXT        NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS trips_rider_id_idx ON trips(rider_id);
