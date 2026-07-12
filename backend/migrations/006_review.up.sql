-- Service: review
-- Reconstructed from backend/services/review/infrastructure/postgres/rating_repository.go
-- CreateSchema() — this is already a production schema-creation function (called
-- from review/cmd/server/main.go at startup), reproduced here verbatim so a
-- fresh database has it before any service starts.

CREATE TABLE IF NOT EXISTS ratings (
    rating_id  TEXT        PRIMARY KEY,
    trip_id    TEXT        NOT NULL,
    rater_id   TEXT        NOT NULL,
    ratee_id   TEXT        NOT NULL,
    role       TEXT        NOT NULL,
    stars      INTEGER     NOT NULL CHECK (stars BETWEEN 1 AND 5),
    comment    TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (trip_id, role)
);
