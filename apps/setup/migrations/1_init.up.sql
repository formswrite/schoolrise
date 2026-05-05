CREATE TABLE setup_state (
    singleton                BOOLEAN PRIMARY KEY DEFAULT TRUE,
    install_token_hash       TEXT,
    install_token_consumed_at TIMESTAMPTZ,
    failed_unlock_attempts   INT NOT NULL DEFAULT 0,
    setup_complete_at        TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT setup_state_singleton CHECK (singleton)
);

INSERT INTO setup_state (singleton) VALUES (TRUE);

CREATE TABLE setup_progress (
    step_code     TEXT PRIMARY KEY,
    payload       JSONB NOT NULL DEFAULT '{}'::jsonb,
    completed_at  TIMESTAMPTZ,
    skipped_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE system_settings (
    key         TEXT PRIMARY KEY,
    value       JSONB NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
