\c setup

CREATE TABLE setup_sessions (
    token_hash  BYTEA PRIMARY KEY,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_setup_sessions_expires ON setup_sessions (expires_at);
