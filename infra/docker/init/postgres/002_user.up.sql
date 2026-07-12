-- Service: user
-- Reconstructed from backend/services/user/infrastructure/postgres/testmain_test.go
-- (createSchema), cross-checked against profile_repository.go — column list
-- (profileFields) matches exactly.

CREATE TABLE IF NOT EXISTS user_profiles (
    id            TEXT        PRIMARY KEY,
    full_name     TEXT        NOT NULL,
    phone         TEXT        NOT NULL,
    email         TEXT        NOT NULL DEFAULT '',
    avatar        TEXT        NOT NULL DEFAULT '',
    date_of_birth TIMESTAMPTZ NULL,
    gender        TEXT        NOT NULL DEFAULT 'unspecified',
    status        TEXT        NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);
