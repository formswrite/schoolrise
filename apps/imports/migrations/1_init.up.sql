CREATE TABLE IF NOT EXISTS imports_health (
    id          BIGSERIAL PRIMARY KEY,
    checked_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
