CREATE TABLE persons (
    id            BIGSERIAL PRIMARY KEY,
    full_name     TEXT NOT NULL,
    given_name    TEXT,
    family_name   TEXT,
    date_of_birth DATE,
    gender        TEXT,
    email         TEXT,
    phone         TEXT,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_persons_full_name ON persons (full_name) WHERE deleted_at IS NULL;
CREATE INDEX idx_persons_email     ON persons (LOWER(email)) WHERE deleted_at IS NULL AND email IS NOT NULL;
