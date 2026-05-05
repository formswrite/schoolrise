-- name: CreateJob :one
INSERT INTO import_jobs (kind, institution_id, status, dry_run, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetJobByID :one
SELECT * FROM import_jobs WHERE id = $1;

-- name: ListJobs :many
SELECT * FROM import_jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateJobResult :one
UPDATE import_jobs
SET status = $2,
    total_rows = $3,
    succeeded_rows = $4,
    failed_rows = $5,
    summary = $6,
    completed_at = now()
WHERE id = $1
RETURNING *;

-- name: AddRowError :exec
INSERT INTO import_row_errors (job_id, row_number, field, error, raw_data)
VALUES ($1, $2, $3, $4, $5);

-- name: ListRowErrors :many
SELECT * FROM import_row_errors
WHERE job_id = $1
ORDER BY row_number, id
LIMIT $2 OFFSET $3;
