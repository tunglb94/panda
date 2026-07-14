-- Auth + Onboarding: Google login resolves accounts by email. A partial
-- unique index (not a plain UNIQUE constraint) because most existing/future
-- phone-only accounts have email = '' and must not collide with each other.
CREATE UNIQUE INDEX IF NOT EXISTS idx_identity_users_email_unique
    ON identity_users (email) WHERE email <> '';
