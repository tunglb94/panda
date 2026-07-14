-- Auth + Onboarding: Rider KYC (Họ tên, Ngày sinh, CCCD, ảnh CCCD trước/sau).
-- One row per rider, created as a draft on first document upload and filled
-- in on submit (see user/domain/entity/rider_verification.go).
CREATE TABLE IF NOT EXISTS rider_verifications (
    id                 TEXT        PRIMARY KEY,
    user_id            TEXT        NOT NULL UNIQUE,
    full_name          TEXT        NOT NULL DEFAULT '',
    date_of_birth      TIMESTAMPTZ NULL,
    national_id_number TEXT        NOT NULL DEFAULT '',
    cccd_front_path    TEXT        NOT NULL DEFAULT '',
    cccd_back_path     TEXT        NOT NULL DEFAULT '',
    status             TEXT        NOT NULL DEFAULT 'pending',
    submitted_at       TIMESTAMPTZ NULL,
    approved_at        TIMESTAMPTZ NULL,
    rejected_at        TIMESTAMPTZ NULL,
    reviewer           TEXT        NOT NULL DEFAULT '',
    reject_reason      TEXT        NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rider_verifications_status ON rider_verifications (status);
