CREATE TABLE academic_periods (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    starts_on   DATE NOT NULL,
    ends_on     DATE NOT NULL,
    is_current  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ,
    CHECK (ends_on >= starts_on)
);

CREATE UNIQUE INDEX idx_academic_periods_code
    ON academic_periods (code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_academic_periods_one_current
    ON academic_periods ((is_current))
    WHERE is_current = TRUE AND deleted_at IS NULL;
