-- Ride Lifecycle Fare Validation hardening.
-- arrived_at/started_at are server-stamped timestamps (MarkDriverArrived/Start)
-- used to derive waiting_duration_min authoritatively, independent of GPS.
-- travelled_distance_km/travelled_duration_min/toll_fee_cents/extra_fee_cents
-- are the Trip Summary — the business-data record Pricing/Settlement/Wallet/
-- Promotion should read instead of raw GPS (see Trip.Complete's doc comment).
ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS arrived_at             TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS started_at              TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS travelled_distance_km   DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS travelled_duration_min  DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS waiting_duration_min    DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS toll_fee_cents          BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS extra_fee_cents         BIGINT NOT NULL DEFAULT 0;
