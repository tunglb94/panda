-- Driver KYC Hardening. Additive only — every change here is a new table,
-- a new nullable/defaulted column, or a new index; no existing column is
-- dropped or renamed, no existing row's meaning changes.

-- ─── Phần 1 — License Capability Rule Engine ───────────────────────────────
-- Replaces the old hardcoded Go map (LicenseClass -> []ServiceType). If
-- Vietnamese law changes which license class permits which service tier,
-- this is a data UPDATE, never a migration/deploy.
CREATE TABLE IF NOT EXISTS license_capabilities (
    license_class TEXT    NOT NULL,
    service_type  TEXT    NOT NULL,
    allowed       BOOLEAN NOT NULL DEFAULT true,
    PRIMARY KEY (license_class, service_type)
);
INSERT INTO license_capabilities (license_class, service_type, allowed) VALUES
    ('A1', 'motorcycle', true),
    ('A1', 'bike_plus',  true),
    ('A2', 'motorcycle', true),
    ('A2', 'bike_plus',  true),
    ('B1', 'car',        true),
    ('B1', 'car_xl',     true),
    ('B2', 'motorcycle', true),
    ('B2', 'bike_plus',  true),
    ('B2', 'car',        true),
    ('B2', 'car_xl',     true)
ON CONFLICT (license_class, service_type) DO NOTHING;

-- ─── Phần 5 — Duplicate Detection (dedup key columns) ──────────────────────
ALTER TABLE driver_verifications ADD COLUMN IF NOT EXISTS national_id_number TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS driver_verifications_national_id_idx
    ON driver_verifications(national_id_number) WHERE national_id_number != '';
CREATE UNIQUE INDEX IF NOT EXISTS driver_verifications_license_number_idx
    ON driver_verifications(license_number) WHERE license_number != '';

-- ─── Phần 6 — Vehicle Identity (optional, unique-if-present) ───────────────
ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS vin            TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS engine_number  TEXT NOT NULL DEFAULT '';
ALTER TABLE vehicle_verifications ADD COLUMN IF NOT EXISTS chassis_number TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS vehicle_verifications_vin_idx
    ON vehicle_verifications(vin) WHERE vin != '';
CREATE UNIQUE INDEX IF NOT EXISTS vehicle_verifications_engine_number_idx
    ON vehicle_verifications(engine_number) WHERE engine_number != '';
CREATE UNIQUE INDEX IF NOT EXISTS vehicle_verifications_chassis_number_idx
    ON vehicle_verifications(chassis_number) WHERE chassis_number != '';

-- ─── Phần 2 + Phần 4 — Document Expiration + Versioning ────────────────────
-- kyc_documents moves from "one row per (driver_id, document_type), upsert
-- on re-upload" to "append-only, one row per uploaded version" — no file or
-- metadata row is ever overwritten, satisfying "Không được mất lịch sử".
ALTER TABLE kyc_documents DROP CONSTRAINT IF EXISTS kyc_documents_driver_id_document_type_key;
ALTER TABLE kyc_documents ADD COLUMN IF NOT EXISTS version     INTEGER     NOT NULL DEFAULT 1;
ALTER TABLE kyc_documents ADD COLUMN IF NOT EXISTS expires_at  TIMESTAMPTZ;
ALTER TABLE kyc_documents ADD COLUMN IF NOT EXISTS uploaded_by TEXT        NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS kyc_documents_driver_type_version_idx
    ON kyc_documents(driver_id, document_type, version);
CREATE INDEX IF NOT EXISTS kyc_documents_expires_at_idx ON kyc_documents(expires_at) WHERE expires_at IS NOT NULL;

-- ─── Phần 7 — Audit Log (append-only; no UPDATE/DELETE ever issued against
-- this table by the application) ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS kyc_audit_logs (
    id          TEXT        PRIMARY KEY,
    entity_type TEXT        NOT NULL, -- 'driver_verification' | 'vehicle_verification' | 'kyc_document'
    entity_id   TEXT        NOT NULL,
    driver_id   TEXT        NOT NULL,
    action      TEXT        NOT NULL, -- 'submit' | 'approve' | 'reject' | 'expire' | 'modify'
    actor_id    TEXT        NOT NULL, -- driver_id, admin user id, or 'system' for auto-expiry
    old_value   TEXT        NOT NULL DEFAULT '',
    new_value   TEXT        NOT NULL DEFAULT '',
    reason      TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS kyc_audit_logs_driver_idx ON kyc_audit_logs(driver_id, created_at DESC);
