-- name: CreatePeriod :one
INSERT INTO academic_periods (code, label, starts_on, ends_on, is_current)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPeriodByID :one
SELECT * FROM academic_periods
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPeriodByCode :one
SELECT * FROM academic_periods
WHERE code = $1 AND deleted_at IS NULL;

-- name: ListPeriods :many
SELECT * FROM academic_periods
WHERE deleted_at IS NULL
ORDER BY starts_on DESC;

-- name: GetCurrentPeriod :one
SELECT * FROM academic_periods
WHERE is_current = TRUE AND deleted_at IS NULL;

-- name: ClearCurrentPeriod :exec
UPDATE academic_periods
SET is_current = FALSE, updated_at = now()
WHERE is_current = TRUE AND deleted_at IS NULL;

-- name: SetPeriodCurrent :one
UPDATE academic_periods
SET is_current = TRUE, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePeriod :exec
UPDATE academic_periods
SET deleted_at = now(), is_current = FALSE, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateNiveau :one
INSERT INTO niveaux (code, label, sort_order)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetNiveauByID :one
SELECT * FROM niveaux
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListNiveaux :many
SELECT * FROM niveaux
WHERE deleted_at IS NULL
ORDER BY sort_order ASC, code ASC;

-- name: SoftDeleteNiveau :exec
UPDATE niveaux
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateClass :one
INSERT INTO classes (period_id, niveau_id, institution_id, code, label, capacity)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetClassByID :one
SELECT * FROM classes
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListClassesByInstitution :many
SELECT * FROM classes
WHERE institution_id = $1 AND deleted_at IS NULL
ORDER BY code ASC;

-- name: ListClassesByPeriod :many
SELECT * FROM classes
WHERE period_id = $1 AND deleted_at IS NULL
ORDER BY institution_id ASC, code ASC;

-- name: SoftDeleteClass :exec
UPDATE classes
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: AddStudentToClass :exec
INSERT INTO class_students (class_id, student_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveStudentFromClass :exec
DELETE FROM class_students
WHERE class_id = $1 AND student_id = $2;

-- name: ListClassStudents :many
SELECT student_id FROM class_students
WHERE class_id = $1
ORDER BY student_id;

-- name: ListClassesForStudent :many
SELECT c.* FROM classes c
JOIN class_students cs ON cs.class_id = c.id
WHERE cs.student_id = $1 AND c.deleted_at IS NULL
ORDER BY c.code;

-- name: ListClassesForAnyStaff :many
SELECT DISTINCT c.* FROM classes c
JOIN class_staff cst ON cst.class_id = c.id
WHERE cst.staff_id = ANY(@staff_ids::bigint[]) AND c.deleted_at IS NULL
ORDER BY c.code;

-- name: IsClassStaff :one
SELECT EXISTS(SELECT 1 FROM class_staff WHERE class_id = $1 AND staff_id = ANY(@staff_ids::bigint[]));

-- name: AddStaffToClass :exec
INSERT INTO class_staff (class_id, staff_id, role)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: RemoveStaffFromClass :exec
DELETE FROM class_staff
WHERE class_id = $1 AND staff_id = $2 AND role = $3;

-- name: ListClassStaff :many
SELECT staff_id, role FROM class_staff
WHERE class_id = $1
ORDER BY role, staff_id;

-- name: ListClassesForStaff :many
SELECT c.* FROM classes c
JOIN class_staff cst ON cst.class_id = c.id
WHERE cst.staff_id = $1 AND c.deleted_at IS NULL
ORDER BY c.code;
