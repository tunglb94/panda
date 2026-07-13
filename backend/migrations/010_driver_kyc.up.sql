-- Driver KYC + Vehicle Verification. Additive only — no existing table
-- touched (driver_profiles keeps its own simpler legacy verification_status
-- column exactly as-is; these new tables are the richer, document-backed,
-- admin-reviewed verification layered on top that actually gates Online).

CREATE TABLE IF NOT EXISTS driver_verifications (
    id              TEXT        PRIMARY KEY,
    driver_id       TEXT        NOT NULL UNIQUE,
    full_name       TEXT        NOT NULL,
    date_of_birth   DATE        NOT NULL,
    address         TEXT        NOT NULL,
    license_number  TEXT        NOT NULL DEFAULT '',
    status          TEXT        NOT NULL DEFAULT 'pending',
    submitted_at    TIMESTAMPTZ NOT NULL,
    approved_at     TIMESTAMPTZ,
    rejected_at     TIMESTAMPTZ,
    reviewer        TEXT        NOT NULL DEFAULT '',
    reject_reason   TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS driver_verifications_status_idx ON driver_verifications(status);

CREATE TABLE IF NOT EXISTS vehicle_verifications (
    id                TEXT        PRIMARY KEY,
    driver_id         TEXT        NOT NULL UNIQUE,
    vehicle_type      TEXT        NOT NULL,
    service_type      TEXT        NOT NULL,
    brand             TEXT        NOT NULL DEFAULT '',
    model             TEXT        NOT NULL DEFAULT '',
    year              INTEGER     NOT NULL DEFAULT 0,
    color             TEXT        NOT NULL DEFAULT '',
    plate_number      TEXT        NOT NULL,
    license_class     TEXT        NOT NULL DEFAULT '',
    ride_enabled      BOOLEAN     NOT NULL DEFAULT false,
    delivery_enabled  BOOLEAN     NOT NULL DEFAULT false,
    status            TEXT        NOT NULL DEFAULT 'pending',
    submitted_at      TIMESTAMPTZ NOT NULL,
    approved_at       TIMESTAMPTZ,
    rejected_at       TIMESTAMPTZ,
    reviewer          TEXT        NOT NULL DEFAULT '',
    reject_reason     TEXT        NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS vehicle_verifications_status_idx ON vehicle_verifications(status);
CREATE INDEX IF NOT EXISTS vehicle_verifications_vehicle_service_idx ON vehicle_verifications(vehicle_type, service_type);
-- A plate can only ever belong to one driver's verification record at a time.
CREATE UNIQUE INDEX IF NOT EXISTS vehicle_verifications_plate_idx ON vehicle_verifications(plate_number) WHERE plate_number != '';

-- kyc_documents is shared by both DriverVerification's documents (CCCD
-- front/back, selfie, license/GPLX) and VehicleVerification's (vehicle
-- registration, insurance) — one row per (driver_id, document_type), upsert
-- on re-upload. storage_path is local-disk-relative and is an internal
-- implementation detail only: the gateway API never echoes it back to a
-- client (see kyc_handler.go) — callers only ever see an opaque document id
-- and an "uploaded: true/false" flag per type. This shape is deliberately
-- generic (no OCR/extracted-fields columns yet) so a later OCR integration
-- (Google Vision/Azure/FPT AI) can add columns here without any schema
-- rework of driver_verifications/vehicle_verifications.
CREATE TABLE IF NOT EXISTS kyc_documents (
    id            TEXT        PRIMARY KEY,
    driver_id     TEXT        NOT NULL,
    document_type TEXT        NOT NULL,
    storage_path  TEXT        NOT NULL,
    content_type  TEXT        NOT NULL DEFAULT '',
    uploaded_at   TIMESTAMPTZ NOT NULL,
    UNIQUE (driver_id, document_type)
);
