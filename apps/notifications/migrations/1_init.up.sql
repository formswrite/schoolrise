CREATE TABLE IF NOT EXISTS notifications_health (
    id          BIGSERIAL PRIMARY KEY,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
