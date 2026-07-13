ALTER TABLE trips DROP COLUMN IF EXISTS commission_rate;
ALTER TABLE trips DROP COLUMN IF EXISTS voucher_discount_cents;
ALTER TABLE trips DROP COLUMN IF EXISTS driver_income_cents;
ALTER TABLE trips DROP COLUMN IF EXISTS commission_cents;
ALTER TABLE trips DROP COLUMN IF EXISTS has_commission_detail;
