\c notifications

CREATE TABLE notifications_outbox (
    id           BIGSERIAL PRIMARY KEY,
    kind         TEXT NOT NULL,
    to_email     TEXT NOT NULL,
    to_name      TEXT,
    subject      TEXT NOT NULL,
    body_html    TEXT NOT NULL,
    body_text    TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending','sending','sent','failed','dropped')),
    attempts     INTEGER NOT NULL DEFAULT 0,
    last_error   TEXT,
    provider_id  TEXT,
    metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_pending
    ON notifications_outbox (scheduled_at)
    WHERE status = 'pending';

CREATE INDEX idx_outbox_kind ON notifications_outbox (kind, created_at DESC);
CREATE INDEX idx_outbox_to ON notifications_outbox (to_email, created_at DESC);
