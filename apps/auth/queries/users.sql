-- name: CreateUser :one
INSERT INTO users (email, password_hash, full_name, role, must_change_password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL;

-- name: GetUserPasswordHash :one
SELECT password_hash FROM users WHERE id = $1;

-- name: IncrementFailedAttempts :exec
UPDATE users
SET failed_attempts = failed_attempts + 1,
    locked_at       = CASE WHEN failed_attempts + 1 >= $2 THEN now() ELSE locked_at END,
    updated_at      = now()
WHERE id = $1;

-- name: ResetFailedAttempts :exec
UPDATE users
SET failed_attempts = 0,
    last_login_at   = now(),
    updated_at      = now()
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY email
LIMIT $1 OFFSET $2;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash        = $2,
    must_change_password = false,
    updated_at           = now()
WHERE id = $1;
