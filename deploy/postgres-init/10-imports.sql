\c imports

CREATE TABLE import_jobs (
    id              BIGSERIAL PRIMARY KEY,
    kind            TEXT NOT NULL,
    institution_id  BIGINT,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','running','completed','failed')),
    total_rows      INTEGER NOT NULL DEFAULT 0,
    succeeded_rows  INTEGER NOT NULL DEFAULT 0,
    failed_rows     INTEGER NOT NULL DEFAULT 0,
    dry_run         BOOLEAN NOT NULL DEFAULT FALSE,
    summary         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_import_jobs_status ON import_jobs (status, created_at DESC);
CREATE INDEX idx_import_jobs_institution ON import_jobs (institution_id) WHERE institution_id IS NOT NULL;

CREATE TABLE import_row_errors (
    id          BIGSERIAL PRIMARY KEY,
    job_id      BIGINT NOT NULL REFERENCES import_jobs(id) ON DELETE CASCADE,
    row_number  INTEGER NOT NULL,
    field       TEXT,
    error       TEXT NOT NULL,
    raw_data    JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_import_row_errors_job ON import_row_errors (job_id, row_number);
