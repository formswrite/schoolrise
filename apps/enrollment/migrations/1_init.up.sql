CREATE TABLE IF NOT EXISTS enrollment_health (
    id          BIGSERIAL PRIMARY KEY,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
