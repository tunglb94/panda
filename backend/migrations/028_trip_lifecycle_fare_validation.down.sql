ALTER TABLE trips
    DROP COLUMN IF EXISTS arrived_at,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS travelled_distance_km,
    DROP COLUMN IF EXISTS travelled_duration_min,
    DROP COLUMN IF EXISTS waiting_duration_min,
    DROP COLUMN IF EXISTS toll_fee_cents,
    DROP COLUMN IF EXISTS extra_fee_cents;
