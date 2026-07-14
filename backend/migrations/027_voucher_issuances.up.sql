-- Voucher & Promotion Hardening, Phase 3 (Rider Voucher Wallet): per-rider
-- issuance, distinct from the voucher_redemptions lifecycle. A rider's
-- wallet only shows vouchers issued to them specifically, not every
-- campaign in existence.
CREATE TABLE IF NOT EXISTS voucher_issuances (
    voucher_id TEXT NOT NULL REFERENCES vouchers(id),
    rider_id   TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'issued',
    issued_at  TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    PRIMARY KEY (voucher_id, rider_id)
);

CREATE INDEX IF NOT EXISTS idx_voucher_issuances_rider ON voucher_issuances(rider_id);
