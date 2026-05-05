-- name: CreateEnrollment :one
INSERT INTO enrollments (student_id, institution_id, period_id, status, enrolled_on)
VALUES ($1, $2, $3, 'active', $4)
RETURNING *;

-- name: GetEnrollmentByID :one
SELECT * FROM enrollments WHERE id = $1;

-- name: GetActiveEnrollment :one
SELECT * FROM enrollments
WHERE student_id = $1 AND period_id = $2 AND status = 'active';

-- name: ListEnrollmentsByInstitution :many
SELECT * FROM enrollments
WHERE institution_id = $1 AND period_id = $2
ORDER BY status, enrolled_on DESC, id DESC;

-- name: ListActiveEnrollmentsByInstitution :many
SELECT * FROM enrollments
WHERE institution_id = $1 AND period_id = $2 AND status = 'active'
ORDER BY enrolled_on DESC, id DESC;

-- name: ListEnrollmentsByStudent :many
SELECT * FROM enrollments
WHERE student_id = $1
ORDER BY enrolled_on DESC, id DESC;

-- name: ListActiveStudentIDsForInstitutions :many
SELECT student_id, institution_id FROM enrollments
WHERE institution_id = ANY(@institution_ids::bigint[])
  AND period_id = $1
  AND status = 'active'
ORDER BY institution_id, student_id;

-- name: SetEnrollmentStatus :one
UPDATE enrollments
SET status = $2, ended_on = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateEnrollmentEvent :one
INSERT INTO enrollment_events (enrollment_id, kind, from_institution_id, to_institution_id, note)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEnrollmentEvents :many
SELECT * FROM enrollment_events
WHERE enrollment_id = $1
ORDER BY occurred_at DESC, id DESC;

-- name: UpsertCoverageSnapshot :one
INSERT INTO coverage_snapshots (scope_node_id, period_id, campaign_id, total_enrolled, total_male, total_female, total_other, total_tested)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetLatestCoverageSnapshot :one
SELECT * FROM coverage_snapshots
WHERE scope_node_id = $1 AND period_id = $2 AND campaign_id IS NOT DISTINCT FROM $3
ORDER BY snapshot_at DESC
LIMIT 1;
