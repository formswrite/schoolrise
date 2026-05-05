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
