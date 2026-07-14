-- Voucher & Promotion Hardening, Phase 1 (Financial Flow) — Trip must
-- persist which voucher (if any) was redeemed against it, so
-- Settlement/Wallet can read it without depending on Promotion being
-- reachable. Additive-only; existing rows default to "no voucher".
ALTER TABLE trips ADD COLUMN voucher_id   TEXT NOT NULL DEFAULT '';
ALTER TABLE trips ADD COLUMN voucher_code TEXT NOT NULL DEFAULT '';
