CREATE TABLE niveaux (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_niveaux_code
    ON niveaux (code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_niveaux_sort
    ON niveaux (sort_order)
    WHERE deleted_at IS NULL;
