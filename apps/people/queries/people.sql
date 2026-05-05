-- name: CreatePerson :one
INSERT INTO persons (full_name, given_name, family_name, date_of_birth, gender, email, phone, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPersonByID :one
SELECT * FROM persons WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeletePerson :exec
UPDATE persons SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateStudent :one
INSERT INTO students (person_id, institution_id, student_code, enrollment_date, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetStudentByID :one
SELECT * FROM students WHERE id = $1 AND deleted_at IS NULL;

-- name: ListStudentsByInstitution :many
SELECT * FROM students
WHERE institution_id = $1 AND deleted_at IS NULL
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: ListStudentsByInstitutionWithPerson :many
SELECT
    s.id              AS student_id,
    s.person_id       AS person_id,
    s.institution_id  AS institution_id,
    s.student_code    AS student_code,
    s.enrollment_date AS enrollment_date,
    s.metadata        AS metadata,
    s.created_at      AS created_at,
    s.updated_at      AS updated_at,
    p.full_name       AS full_name,
    p.given_name      AS given_name,
    p.family_name     AS family_name,
    p.date_of_birth   AS date_of_birth,
    p.gender          AS gender,
    p.email           AS email,
    p.phone           AS phone,
    p.metadata        AS person_metadata
FROM students s
JOIN persons p ON p.id = s.person_id
WHERE s.institution_id = $1
  AND s.deleted_at IS NULL
  AND p.deleted_at IS NULL
ORDER BY s.id
LIMIT $2 OFFSET $3;

-- name: ListStudentsByInstitutions :many
SELECT * FROM students
WHERE institution_id = ANY($1::bigint[]) AND deleted_at IS NULL
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: SoftDeleteStudent :exec
UPDATE students SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPersonByEmail :one
SELECT * FROM persons
WHERE lower(email) = lower($1) AND deleted_at IS NULL
LIMIT 1;

-- name: ListStaffByPersonID :many
SELECT * FROM staff
WHERE person_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: GetStudentGendersByIDs :many
SELECT s.id AS student_id, p.gender
FROM students s
JOIN persons p ON p.id = s.person_id
WHERE s.id = ANY(@student_ids::bigint[])
  AND s.deleted_at IS NULL
  AND p.deleted_at IS NULL;

-- name: GetSchoolsForStudents :many
SELECT id AS student_id, institution_id
FROM students
WHERE id = ANY(@student_ids::bigint[]) AND deleted_at IS NULL;

-- name: CreateStaff :one
INSERT INTO staff (person_id, scope_node_id, position, staff_code, hire_date, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetStaffByID :one
SELECT * FROM staff WHERE id = $1 AND deleted_at IS NULL;

-- name: ListStaffByScope :many
SELECT * FROM staff
WHERE scope_node_id = $1 AND deleted_at IS NULL
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: ListStaffByScopes :many
SELECT * FROM staff
WHERE scope_node_id = ANY($1::bigint[]) AND deleted_at IS NULL
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: SoftDeleteStaff :exec
UPDATE staff SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;
