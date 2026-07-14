-- Promotion Engine: the first real migration for this service — until now
-- backend/services/promotion/infrastructure/postgres/testmain_test.go's
-- createSchema was the only place this schema existed ("test scaffolding
-- only, NOT a migration" per that file's own comment). This migration
-- matches that test schema exactly, plus service_types/trip_types/campaign
-- (Voucher & Promotion production-ready phase).
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
CREATE INDEX IF NOT EXISTS vouchers_campaign_idx ON vouchers (campaign) WHERE campaign <> '';

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
CREATE INDEX IF NOT EXISTS voucher_redemptions_rider_idx ON voucher_redemptions(rider_id, created_at DESC);
