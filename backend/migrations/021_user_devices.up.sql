-- Device registry — upserted on every successful login (see
-- identity/app/upsert_device.go). One row per (user_id, device_id); a user
-- with multiple devices (phone + tablet) gets multiple rows.
CREATE TABLE IF NOT EXISTS identity_user_devices (
    user_id     TEXT        NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    device_id   TEXT        NOT NULL,
    platform    TEXT        NOT NULL DEFAULT '',
    model       TEXT        NOT NULL DEFAULT '',
    app_version TEXT        NOT NULL DEFAULT '',
    fcm_token   TEXT        NOT NULL DEFAULT '',
    last_seen   TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, device_id)
);
