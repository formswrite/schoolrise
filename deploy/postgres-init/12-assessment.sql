\c assessment

CREATE TABLE scales (
    code        TEXT PRIMARY KEY,
    label       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE scale_bands (
    id          BIGSERIAL PRIMARY KEY,
    scale_code  TEXT NOT NULL REFERENCES scales(code) ON DELETE CASCADE,
    ordinal     INTEGER NOT NULL,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    min_score   INTEGER NOT NULL DEFAULT 0,
    max_score   INTEGER NOT NULL DEFAULT 0,
    UNIQUE (scale_code, ordinal),
    UNIQUE (scale_code, code)
);

CREATE INDEX idx_scale_bands_scale ON scale_bands (scale_code, ordinal);

INSERT INTO scales (code, label) VALUES
    ('french_5level', 'Compétences en Français'),
    ('maths_5level',  'Compétences en Mathématiques');

INSERT INTO scale_bands (scale_code, ordinal, code, label, min_score, max_score) VALUES
    ('french_5level', 1, 'debutant',   'Débutant',   0,  19),
    ('french_5level', 2, 'lettres',    'Lettres',    20, 39),
    ('french_5level', 3, 'mots',       'Mots',       40, 59),
    ('french_5level', 4, 'paragraphe', 'Paragraphe', 60, 79),
    ('french_5level', 5, 'histoire',   'Histoire',   80, 100),
    ('maths_5level',  1, 'debutant',     'Débutant',     0,  19),
    ('maths_5level',  2, 'un_chiffre',   '1 chiffre',    20, 39),
    ('maths_5level',  3, 'deux_chiffres','2 chiffres',   40, 59),
    ('maths_5level',  4, 'soustraction', 'Soustraction', 60, 79),
    ('maths_5level',  5, 'division',     'Division',     80, 100);

CREATE TABLE campaigns (
    id              BIGSERIAL PRIMARY KEY,
    public_id       TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    scale_code      TEXT NOT NULL REFERENCES scales(code),
    form_id         BIGINT NOT NULL,
    form_version_id BIGINT NOT NULL,
    period_id       BIGINT NOT NULL,
    scope_node_id   BIGINT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','open','closed')),
    opens_at        TIMESTAMPTZ,
    closes_at       TIMESTAMPTZ,
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_campaigns_scope ON campaigns (scope_node_id, status);
CREATE INDEX idx_campaigns_form_version ON campaigns (form_version_id);

CREATE TABLE assignments (
    id            BIGSERIAL PRIMARY KEY,
    campaign_id   BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    student_id    BIGINT NOT NULL,
    access_token  TEXT NOT NULL UNIQUE,
    submitted_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (campaign_id, student_id)
);

CREATE INDEX idx_assignments_campaign ON assignments (campaign_id);

CREATE TABLE responses (
    id                    BIGSERIAL PRIMARY KEY,
    assignment_id         BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    campaign_id           BIGINT NOT NULL,
    student_id            BIGINT NOT NULL,
    payload               JSONB NOT NULL DEFAULT '{}'::jsonb,
    submitted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    proctored_by_user_id  BIGINT,
    entry_mode            TEXT NOT NULL DEFAULT 'student'
                          CHECK (entry_mode IN ('student','proctored_score','proctored_answers'))
);

CREATE INDEX idx_responses_campaign ON responses (campaign_id);
CREATE INDEX idx_responses_student ON responses (student_id);
CREATE INDEX idx_responses_proctored_by ON responses (proctored_by_user_id) WHERE proctored_by_user_id IS NOT NULL;

CREATE TABLE scores (
    id              BIGSERIAL PRIMARY KEY,
    response_id     BIGINT NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
    campaign_id     BIGINT NOT NULL,
    student_id      BIGINT NOT NULL,
    raw_score       INTEGER NOT NULL,
    band_code       TEXT NOT NULL,
    band_ordinal    INTEGER NOT NULL,
    finalized_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (response_id)
);

CREATE INDEX idx_scores_campaign ON scores (campaign_id);
CREATE INDEX idx_scores_student ON scores (student_id);
CREATE UNIQUE INDEX idx_scores_unique_per_student_campaign ON scores (student_id, campaign_id);
