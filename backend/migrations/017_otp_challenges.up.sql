-- Auth + Onboarding: phone OTP login. Codes are never stored plaintext —
-- only a salted SHA-256 hash (see identity/domain/entity/otp_challenge.go).
CREATE TABLE IF NOT EXISTS identity_otp_challenges (
    id           TEXT        PRIMARY KEY,
    phone_number TEXT        NOT NULL,
    code_hash    TEXT        NOT NULL,
    purpose      TEXT        NOT NULL DEFAULT 'login',
    expires_at   TIMESTAMPTZ NOT NULL,
    attempts     INT         NOT NULL DEFAULT 0,
    consumed     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_identity_otp_challenges_phone
    ON identity_otp_challenges (phone_number, created_at DESC);
