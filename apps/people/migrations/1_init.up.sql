CREATE TABLE IF NOT EXISTS people_health (
    id          BIGSERIAL PRIMARY KEY,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
