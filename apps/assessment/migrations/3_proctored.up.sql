ALTER TABLE responses
    ADD COLUMN proctored_by_user_id BIGINT,
    ADD COLUMN entry_mode TEXT NOT NULL DEFAULT 'student'
        CHECK (entry_mode IN ('student','proctored_score','proctored_answers'));

CREATE INDEX idx_responses_proctored_by ON responses (proctored_by_user_id) WHERE proctored_by_user_id IS NOT NULL;

CREATE UNIQUE INDEX idx_scores_unique_per_student_campaign ON scores (student_id, campaign_id);
