CREATE TABLE students (
    id              BIGSERIAL PRIMARY KEY,
    person_id       BIGINT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    institution_id  BIGINT NOT NULL,
    student_code    TEXT,
    enrollment_date DATE,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_students_person ON students (person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_students_institution ON students (institution_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_students_code_per_institution
    ON students (institution_id, student_code)
    WHERE deleted_at IS NULL AND student_code IS NOT NULL;
