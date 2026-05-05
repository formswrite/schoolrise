\c people

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

CREATE TABLE staff (
    id            BIGSERIAL PRIMARY KEY,
    person_id     BIGINT NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    scope_node_id BIGINT NOT NULL,
    position      TEXT,
    staff_code    TEXT,
    hire_date     DATE,
    metadata      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_staff_person ON staff (person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_staff_scope ON staff (scope_node_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_staff_code_per_scope
    ON staff (scope_node_id, staff_code)
    WHERE deleted_at IS NULL AND staff_code IS NOT NULL;
