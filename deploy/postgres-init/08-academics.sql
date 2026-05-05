\c academics

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
    ON academic_periods (code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_academic_periods_one_current
    ON academic_periods ((is_current))
    WHERE is_current = TRUE AND deleted_at IS NULL;

CREATE TABLE niveaux (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_niveaux_code ON niveaux (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_niveaux_sort ON niveaux (sort_order) WHERE deleted_at IS NULL;

CREATE TABLE classes (
    id              BIGSERIAL PRIMARY KEY,
    period_id       BIGINT NOT NULL REFERENCES academic_periods(id) ON DELETE RESTRICT,
    niveau_id       BIGINT NOT NULL REFERENCES niveaux(id) ON DELETE RESTRICT,
    institution_id  BIGINT NOT NULL,
    code            TEXT NOT NULL,
    label           TEXT NOT NULL,
    capacity        INTEGER,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_classes_unique
    ON classes (period_id, institution_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_classes_institution ON classes (institution_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_classes_niveau ON classes (niveau_id) WHERE deleted_at IS NULL;

CREATE TABLE class_students (
    class_id    BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id  BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (class_id, student_id)
);

CREATE INDEX idx_class_students_student ON class_students (student_id);

CREATE TABLE class_staff (
    class_id    BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    staff_id    BIGINT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'teacher',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (class_id, staff_id, role)
);

CREATE INDEX idx_class_staff_staff ON class_staff (staff_id);
