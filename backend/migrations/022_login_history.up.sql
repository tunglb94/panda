-- Append-only login audit trail. No foreign key on user_id — a failed OTP
-- attempt (wrong code, unknown phone) is still recorded but may not have
-- resolved to a real account yet, so user_id can be ''.
CREATE TABLE IF NOT EXISTS identity_login_history (
    id           TEXT        PRIMARY KEY,
    user_id      TEXT        NOT NULL DEFAULT '',
    login_time   TIMESTAMPTZ NOT NULL,
    ip           TEXT        NOT NULL DEFAULT '',
    device_id    TEXT        NOT NULL DEFAULT '',
    platform     TEXT        NOT NULL DEFAULT '',
    login_method TEXT        NOT NULL DEFAULT '',
    success      BOOLEAN     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_identity_login_history_user_id
    ON identity_login_history (user_id, login_time DESC);
