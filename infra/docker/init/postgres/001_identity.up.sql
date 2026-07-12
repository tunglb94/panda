-- Service: identity
-- Reconstructed from backend/services/identity/infrastructure/postgres/testmain_test.go
-- (createSchema) and cross-checked against permission_repository.go, role_repository.go,
-- user_repository.go — all column lists match the test scaffolding exactly.

CREATE TABLE IF NOT EXISTS identity_permissions (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    resource    TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS identity_roles (
    id          TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS identity_role_permissions (
    role_id       TEXT NOT NULL REFERENCES identity_roles(id)       ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES identity_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS identity_users (
    id           TEXT        PRIMARY KEY,
    phone_number TEXT        NOT NULL UNIQUE,
    name         TEXT        NOT NULL,
    email        TEXT        NOT NULL DEFAULT '',
    type         TEXT        NOT NULL,
    status       TEXT        NOT NULL,
    role_id      TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL
);
