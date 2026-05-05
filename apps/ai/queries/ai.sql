-- name: CreateJob :one
INSERT INTO ai_jobs (kind, model, prompt_summary, requested_by, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CompleteJob :exec
UPDATE ai_jobs
SET status = 'done',
    completed_at = now(),
    request_tokens = $2,
    response_tokens = $3,
    latency_ms = $4
WHERE id = $1;

-- name: FailJob :exec
UPDATE ai_jobs
SET status = 'failed',
    completed_at = now(),
    error = $2,
    latency_ms = $3
WHERE id = $1;

-- name: ListRecentJobs :many
SELECT * FROM ai_jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetJobByID :one
SELECT * FROM ai_jobs WHERE id = $1;
