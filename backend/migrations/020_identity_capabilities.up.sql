-- Auth/KYC hardening: users are no longer one-account-per-type. Every
-- self-registered account defaults to Rider; driver_enabled is a capability
-- flag so the same account can also use the Driver app (see
-- identity/domain/entity/user.go). phone_number keeps its existing
-- NOT NULL DEFAULT '' convention (matches the pre-existing `email` column)
-- rather than becoming nullable — "not set" is empty string, not NULL,
-- everywhere on this row.
ALTER TABLE identity_users DROP CONSTRAINT IF EXISTS identity_users_phone_number_key;
ALTER TABLE identity_users ALTER COLUMN phone_number SET DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_phone_unique
    ON identity_users (phone_number) WHERE phone_number <> '';

ALTER TABLE identity_users ADD COLUMN IF NOT EXISTS google_sub TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_google_sub_unique
    ON identity_users (google_sub) WHERE google_sub <> '';

ALTER TABLE identity_users ADD COLUMN IF NOT EXISTS driver_enabled BOOLEAN NOT NULL DEFAULT FALSE;
