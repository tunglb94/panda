-- Fixes a pre-existing gap: driver/infrastructure/postgres's Go code has
-- always selected an `expired_at` column for both driver_verifications and
-- vehicle_verifications (Document Expiration / KYC Hardening phase), but no
-- migration ever actually created it — every GET on these tables failed
-- with "column expired_at does not exist" (surfaced as a generic 500
-- "scan failed" once pgx defers the query error to Scan()).
ALTER TABLE driver_verifications ADD COLUMN expired_at TIMESTAMPTZ;
ALTER TABLE vehicle_verifications ADD COLUMN expired_at TIMESTAMPTZ;
