-- name: CreateForm :one
INSERT INTO forms (public_id, owner_id, title, description, settings)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFormByID :one
SELECT * FROM forms WHERE id = $1 AND deleted_at IS NULL;

-- name: GetFormByPublicID :one
SELECT * FROM forms WHERE public_id = $1 AND deleted_at IS NULL;

-- name: ListFormsByOwner :many
SELECT * FROM forms
WHERE owner_id = $1 AND deleted_at IS NULL
ORDER BY updated_at DESC, id DESC;

-- name: UpdateFormMeta :one
UPDATE forms
SET title = $2, description = $3, settings = $4, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateFormStatus :one
UPDATE forms
SET status = $2, updated_at = now(),
    published_at = CASE WHEN $2 = 'published' THEN now() ELSE published_at END
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteForm :exec
UPDATE forms SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateQuestion :one
INSERT INTO questions (form_id, client_id, sort_order, title, description, type, required, options, scale_min, scale_max, scale_labels, validation, grading, extra)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1 AND deleted_at IS NULL;

-- name: ListQuestionsByForm :many
SELECT * FROM questions
WHERE form_id = $1 AND deleted_at IS NULL
ORDER BY sort_order ASC, id ASC;

-- name: UpdateQuestion :one
UPDATE questions
SET title = $2, description = $3, type = $4, required = $5,
    sort_order = $6, options = $7, scale_min = $8, scale_max = $9, scale_labels = $10,
    validation = $11, grading = $12, extra = $13, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteQuestion :exec
UPDATE questions SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateFormVersion :one
INSERT INTO form_versions (form_id, version_num, title, description, snapshot)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLatestFormVersion :one
SELECT * FROM form_versions
WHERE form_id = $1
ORDER BY version_num DESC
LIMIT 1;

-- name: GetFormVersion :one
SELECT * FROM form_versions WHERE id = $1;

-- name: ListFormVersions :many
SELECT * FROM form_versions WHERE form_id = $1 ORDER BY version_num DESC;
