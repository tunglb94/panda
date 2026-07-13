-- Voucher pending marker (Fix #2) — distinguishes a confirmed "no voucher"
-- from "Promotion Engine detail unavailable at settlement time" instead of
-- both silently recording voucher_cost_cents = 0. Existing rows predate
-- Pricing V3 detail ever reaching Settlement, hence the 'unknown' default.
ALTER TABLE settlements ADD COLUMN voucher_status VARCHAR(16) NOT NULL DEFAULT 'unknown';
