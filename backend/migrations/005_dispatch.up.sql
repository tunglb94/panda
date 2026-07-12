-- Service: dispatch
-- Reconstructed from backend/services/dispatch/infrastructure/postgres/testmain_test.go
-- (createSchema), cross-checked against dispatch_repository.go / transactor.go —
-- column list matches exactly. The partial index is copied verbatim from the
-- test schema (used by FindExpiredOffers' WHERE status='searching' AND
-- offer_expires_at < $1 query).
--
-- dispatch's own test schema also declares a smaller/divergent local copy of
-- `trips` for its own package-local integration tests. That copy is NOT
-- reproduced here: dispatch's only real production writes to `trips` (via
-- TripUpdater, see trip_updater.go) are plain UPDATEs of status/driver_id/
-- updated_at against the trips table owned by 004_trip.up.sql, which already
-- satisfies every column dispatch touches. Reproducing dispatch's reduced
-- local schema here would conflict with (and be redundant with) 004.

CREATE TABLE IF NOT EXISTS dispatch_jobs (
    job_id             TEXT             PRIMARY KEY,
    trip_id            TEXT             NOT NULL UNIQUE,
    rider_id           TEXT             NOT NULL,
    pickup_lat         DOUBLE PRECISION NOT NULL,
    pickup_lon         DOUBLE PRECISION NOT NULL,
    status             TEXT             NOT NULL DEFAULT 'pending',
    current_driver_id  TEXT             NOT NULL DEFAULT '',
    assigned_driver_id TEXT             NOT NULL DEFAULT '',
    offered_driver_ids TEXT             NOT NULL DEFAULT '',
    offer_expires_at   TIMESTAMPTZ,
    offer_timeout_sec  INT              NOT NULL DEFAULT 30,
    max_attempts       INT              NOT NULL DEFAULT 5,
    attempt_count      INT              NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ      NOT NULL,
    updated_at         TIMESTAMPTZ      NOT NULL
);

CREATE INDEX IF NOT EXISTS dispatch_jobs_status_expires_idx
    ON dispatch_jobs(offer_expires_at)
    WHERE status = 'searching';
