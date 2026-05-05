\c ai

CREATE TABLE ai_jobs (
    id                 BIGSERIAL PRIMARY KEY,
    kind               TEXT NOT NULL
                       CHECK (kind IN ('suggest_items','draft_rubric','grade_essay','generate_distractors')),
    model              TEXT NOT NULL DEFAULT '',
    status             TEXT NOT NULL DEFAULT 'pending'
                       CHECK (status IN ('pending','running','done','failed')),
    prompt_summary     TEXT NOT NULL DEFAULT '',
    request_tokens     INTEGER NOT NULL DEFAULT 0,
    response_tokens    INTEGER NOT NULL DEFAULT 0,
    latency_ms         INTEGER NOT NULL DEFAULT 0,
    error              TEXT,
    requested_by       BIGINT,
    metadata           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at       TIMESTAMPTZ
);

CREATE INDEX idx_ai_jobs_kind     ON ai_jobs (kind, created_at DESC);
CREATE INDEX idx_ai_jobs_status   ON ai_jobs (status, created_at DESC);
CREATE INDEX idx_ai_jobs_requester ON ai_jobs (requested_by, created_at DESC) WHERE requested_by IS NOT NULL;
