-- Communication Module: in-app chat + call audit log + in-app notifications.
-- Additive only — no existing table touched. All 4 tables are owned by the
-- "notification" service's domain model (backend/services/notification),
-- whose logic runs in-process inside Gateway (see module report, "Kien truc"
-- — no protoc/buf toolchain is available in this environment to generate a
-- brand-new gRPC proto package, so Gateway imports the notification Go
-- packages directly against this same database, mirroring the existing
-- identity-service precedent already in this codebase).

CREATE TABLE IF NOT EXISTS conversations (
    id         TEXT        PRIMARY KEY,
    trip_id    TEXT        NOT NULL UNIQUE,
    rider_id   TEXT        NOT NULL,
    driver_id  TEXT        NOT NULL,
    trip_type  TEXT        NOT NULL DEFAULT 'ride',
    status     TEXT        NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL,
    closed_at  TIMESTAMPTZ
);

-- messages.seq is a monotonic per-database ordering cursor used as the
-- "since_id" polling cursor by ListMessages/PollMessages — the text `id`
-- (UUID) stays the externally-facing identifier, consistent with every
-- other aggregate in this codebase, but UUIDs don't sort chronologically so
-- a separate serial column is needed for cheap "give me everything after X" queries.
CREATE TABLE IF NOT EXISTS messages (
    seq             BIGSERIAL   PRIMARY KEY,
    id              TEXT        NOT NULL UNIQUE,
    conversation_id TEXT        NOT NULL REFERENCES conversations(id),
    sender_id       TEXT        NOT NULL,
    sender_role     TEXT        NOT NULL,
    body            TEXT        NOT NULL,
    quick_reply_key TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS messages_conversation_seq_idx ON messages(conversation_id, seq);

CREATE TABLE IF NOT EXISTS notifications (
    id              TEXT        PRIMARY KEY,
    user_id         TEXT        NOT NULL,
    category        TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    body            TEXT        NOT NULL DEFAULT '',
    trip_id         TEXT        NOT NULL DEFAULT '',
    conversation_id TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL,
    read_at         TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS notifications_user_created_idx ON notifications(user_id, created_at DESC);

-- call_sessions is an audit log only (Part 1 — "CallSession được ghi log cho
-- mục audit, không phải để phát lại số"); no real telephony/virtual-number
-- integration exists, so there is no call duration/status to record yet.
CREATE TABLE IF NOT EXISTS call_sessions (
    id         TEXT        PRIMARY KEY,
    trip_id    TEXT        NOT NULL,
    caller_id  TEXT        NOT NULL,
    callee_id  TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS call_sessions_trip_id_idx ON call_sessions(trip_id);
