-- Trip commission detail (Settlement Fix #1) — persists the commission
-- Pricing V3 actually computed for this trip's final fare, so Settlement
-- reads it instead of inventing its own flat rate. Additive-only; existing
-- rows default to "no detail available" (has_commission_detail = false).
ALTER TABLE trips ADD COLUMN has_commission_detail   BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE trips ADD COLUMN commission_cents        BIGINT NOT NULL DEFAULT 0;
ALTER TABLE trips ADD COLUMN driver_income_cents      BIGINT NOT NULL DEFAULT 0;
ALTER TABLE trips ADD COLUMN voucher_discount_cents   BIGINT NOT NULL DEFAULT 0;
ALTER TABLE trips ADD COLUMN commission_rate          DOUBLE PRECISION NOT NULL DEFAULT 0;
