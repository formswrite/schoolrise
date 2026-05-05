-- name: ListScales :many
SELECT * FROM scales ORDER BY code;

-- name: GetScale :one
SELECT * FROM scales WHERE code = $1;

-- name: ListBandsForScale :many
SELECT * FROM scale_bands WHERE scale_code = $1 ORDER BY ordinal ASC;

-- name: BandForScore :one
SELECT * FROM scale_bands
WHERE scale_code = $1 AND $2::int >= min_score AND $2::int <= max_score
ORDER BY ordinal DESC
LIMIT 1;

-- name: CreateCampaign :one
INSERT INTO campaigns (public_id, title, scale_code, form_id, form_version_id, period_id, scope_node_id, status, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1;

-- name: GetCampaignByPublicID :one
SELECT * FROM campaigns WHERE public_id = $1;

-- name: ListCampaignsByScope :many
SELECT * FROM campaigns
WHERE scope_node_id = $1
ORDER BY created_at DESC;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET status = $2,
    opens_at = CASE WHEN $2 = 'open' AND opens_at IS NULL THEN now() ELSE opens_at END,
    closes_at = CASE WHEN $2 = 'closed' THEN now() ELSE closes_at END,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateAssignment :one
INSERT INTO assignments (campaign_id, student_id, access_token)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetAssignmentByToken :one
SELECT * FROM assignments WHERE access_token = $1;

-- name: ListAssignmentsByCampaign :many
SELECT * FROM assignments WHERE campaign_id = $1 ORDER BY id;

-- name: MarkAssignmentSubmitted :one
UPDATE assignments
SET submitted_at = now()
WHERE id = $1 AND submitted_at IS NULL
RETURNING *;

-- name: CreateResponse :one
INSERT INTO responses (assignment_id, campaign_id, student_id, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListResponsesByCampaign :many
SELECT * FROM responses WHERE campaign_id = $1 ORDER BY submitted_at DESC;

-- name: CreateScore :one
INSERT INTO scores (response_id, campaign_id, student_id, raw_score, band_code, band_ordinal)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetScoreByResponse :one
SELECT * FROM scores WHERE response_id = $1;

-- name: ListScoresByCampaign :many
SELECT * FROM scores WHERE campaign_id = $1 ORDER BY band_ordinal DESC, raw_score DESC;

-- name: GetAssignmentByCampaignStudent :one
SELECT * FROM assignments
WHERE campaign_id = $1 AND student_id = $2;

-- name: DeleteResponseByAssignment :exec
DELETE FROM responses WHERE assignment_id = $1;

-- name: CreateProctoredResponse :one
INSERT INTO responses (assignment_id, campaign_id, student_id, payload, proctored_by_user_id, entry_mode)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListGradingRosterRows :many
SELECT
    cs.student_id::bigint        AS student_id,
    a.id                         AS assignment_id,
    s.id                         AS score_id,
    s.raw_score                  AS raw_score,
    s.band_code                  AS band_code,
    s.band_ordinal               AS band_ordinal,
    r.entry_mode                 AS entry_mode,
    r.submitted_at               AS submitted_at,
    r.proctored_by_user_id       AS proctored_by_user_id
FROM unnest(@student_ids::bigint[]) AS cs(student_id)
LEFT JOIN assignments a ON a.campaign_id = $1 AND a.student_id = cs.student_id
LEFT JOIN scores s      ON s.campaign_id = $1 AND s.student_id = cs.student_id
LEFT JOIN responses r   ON r.id = (SELECT id FROM responses WHERE campaign_id = $1 AND student_id = cs.student_id ORDER BY id DESC LIMIT 1)
ORDER BY cs.student_id;

-- name: ListOpenCampaignsWithScores :many
SELECT DISTINCT c.id, c.period_id, c.scope_node_id, c.scale_code
FROM campaigns c
JOIN scores s ON s.campaign_id = c.id
WHERE c.status = 'open';

-- name: ListScoredStudentIDs :many
SELECT DISTINCT student_id FROM scores WHERE campaign_id = $1;
