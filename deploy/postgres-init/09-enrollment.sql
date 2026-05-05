\c enrollment

CREATE TABLE enrollments (
    id              BIGSERIAL PRIMARY KEY,
    student_id      BIGINT NOT NULL,
    institution_id  BIGINT NOT NULL,
    period_id       BIGINT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','transferred','dropped','graduated','reinstated')),
    enrolled_on     DATE NOT NULL,
    ended_on        DATE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_enrollments_active_unique
    ON enrollments (student_id, period_id) WHERE status = 'active';
CREATE INDEX idx_enrollments_institution ON enrollments (institution_id, period_id);
CREATE INDEX idx_enrollments_student ON enrollments (student_id);

CREATE TABLE enrollment_events (
    id                   BIGSERIAL PRIMARY KEY,
    enrollment_id        BIGINT NOT NULL REFERENCES enrollments(id) ON DELETE CASCADE,
    kind                 TEXT NOT NULL
                         CHECK (kind IN ('created','transferred','dropped','graduated','reinstated')),
    from_institution_id  BIGINT,
    to_institution_id    BIGINT,
    note                 TEXT,
    occurred_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_enrollment_events_enrollment
    ON enrollment_events (enrollment_id, occurred_at DESC);

CREATE TABLE coverage_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    scope_node_id   BIGINT NOT NULL,
    period_id       BIGINT NOT NULL,
    campaign_id     BIGINT,
    total_enrolled  INTEGER NOT NULL DEFAULT 0,
    total_male      INTEGER NOT NULL DEFAULT 0,
    total_female    INTEGER NOT NULL DEFAULT 0,
    total_other     INTEGER NOT NULL DEFAULT 0,
    total_tested    INTEGER NOT NULL DEFAULT 0,
    snapshot_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_coverage_snapshots_scope_period
    ON coverage_snapshots (scope_node_id, period_id, campaign_id);
