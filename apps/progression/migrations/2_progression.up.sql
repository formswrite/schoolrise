CREATE TABLE progression_snapshots (
    id            BIGSERIAL PRIMARY KEY,
    scope_node_id BIGINT NOT NULL,
    period_id     BIGINT NOT NULL,
    campaign_id   BIGINT NOT NULL,
    band_code     TEXT NOT NULL,
    band_ordinal  INTEGER NOT NULL,
    student_count INTEGER NOT NULL DEFAULT 0,
    snapshot_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (scope_node_id, period_id, campaign_id, band_code)
);

CREATE INDEX idx_progression_snapshots_scope
    ON progression_snapshots (scope_node_id, period_id, campaign_id);

CREATE TABLE progression_refresh_log (
    id            BIGSERIAL PRIMARY KEY,
    campaign_id   BIGINT NOT NULL,
    triggered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,
    rows_written  INTEGER NOT NULL DEFAULT 0,
    error         TEXT
);

CREATE INDEX idx_progression_refresh_log_campaign
    ON progression_refresh_log (campaign_id, triggered_at DESC);
