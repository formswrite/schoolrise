CREATE TABLE forms (
    id            BIGSERIAL PRIMARY KEY,
    public_id     TEXT NOT NULL UNIQUE,
    owner_id      BIGINT NOT NULL,
    title         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'draft'
                  CHECK (status IN ('draft','published','closed')),
    settings      JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_count INTEGER NOT NULL DEFAULT 0,
    view_count     INTEGER NOT NULL DEFAULT 0,
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_forms_owner ON forms (owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_forms_status ON forms (status) WHERE deleted_at IS NULL;

CREATE TABLE form_versions (
    id           BIGSERIAL PRIMARY KEY,
    form_id      BIGINT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    version_num  INTEGER NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    snapshot     JSONB NOT NULL,
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (form_id, version_num)
);

CREATE INDEX idx_form_versions_form ON form_versions (form_id);

CREATE TABLE questions (
    id            BIGSERIAL PRIMARY KEY,
    form_id       BIGINT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    client_id     TEXT NOT NULL,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    title         TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    type          TEXT NOT NULL,
    required      BOOLEAN NOT NULL DEFAULT FALSE,
    options       JSONB NOT NULL DEFAULT '[]'::jsonb,
    scale_min     INTEGER,
    scale_max     INTEGER,
    scale_labels  JSONB NOT NULL DEFAULT '{}'::jsonb,
    validation    JSONB NOT NULL DEFAULT '{}'::jsonb,
    grading       JSONB NOT NULL DEFAULT '{}'::jsonb,
    extra         JSONB NOT NULL DEFAULT '{}'::jsonb,
    image         JSONB NOT NULL DEFAULT '{}'::jsonb,
    translations  JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    UNIQUE (form_id, client_id)
);

CREATE INDEX idx_questions_form_order ON questions (form_id, sort_order) WHERE deleted_at IS NULL;

CREATE TABLE form_logic_rules (
    id              BIGSERIAL PRIMARY KEY,
    form_id         BIGINT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    target_question_client_id TEXT NOT NULL,
    operator        TEXT NOT NULL CHECK (operator IN ('show_if','hide_if')),
    conditions      JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_form_logic_rules_form ON form_logic_rules (form_id);
