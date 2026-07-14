ALTER TABLE identity_users DROP COLUMN IF EXISTS driver_enabled;

DROP INDEX IF EXISTS idx_identity_users_google_sub_unique;
ALTER TABLE identity_users DROP COLUMN IF EXISTS google_sub;

DROP INDEX IF EXISTS idx_identity_users_phone_unique;
ALTER TABLE identity_users ADD CONSTRAINT identity_users_phone_number_key UNIQUE (phone_number);
