-- name: GetSetupState :one
SELECT install_token_hash, install_token_consumed_at, failed_unlock_attempts,
       setup_complete_at, created_at, updated_at
FROM setup_state WHERE singleton = TRUE;

-- name: SetInstallTokenHash :exec
UPDATE setup_state
SET install_token_hash = $1,
    updated_at         = now()
WHERE singleton = TRUE;

-- name: ConsumeInstallToken :exec
UPDATE setup_state
SET install_token_consumed_at = now(),
    updated_at                = now()
WHERE singleton = TRUE
  AND install_token_consumed_at IS NULL;

-- name: IncrementFailedUnlockAttempts :exec
UPDATE setup_state
SET failed_unlock_attempts = failed_unlock_attempts + 1,
    updated_at             = now()
WHERE singleton = TRUE;

-- name: ResetFailedUnlockAttempts :exec
UPDATE setup_state
SET failed_unlock_attempts = 0,
    updated_at             = now()
WHERE singleton = TRUE;

-- name: MarkSetupComplete :exec
UPDATE setup_state
SET setup_complete_at = now(),
    updated_at        = now()
WHERE singleton = TRUE
  AND setup_complete_at IS NULL;

-- name: ListSetupProgress :many
SELECT step_code, payload, completed_at, skipped_at, created_at, updated_at
FROM setup_progress
ORDER BY step_code;

-- name: GetSetupProgressStep :one
SELECT step_code, payload, completed_at, skipped_at, created_at, updated_at
FROM setup_progress WHERE step_code = $1;

-- name: UpsertSetupProgressComplete :exec
INSERT INTO setup_progress (step_code, payload, completed_at)
VALUES ($1, $2, now())
ON CONFLICT (step_code) DO UPDATE
SET payload      = EXCLUDED.payload,
    completed_at = now(),
    skipped_at   = NULL,
    updated_at   = now();

-- name: UpsertSetupProgressSkip :exec
INSERT INTO setup_progress (step_code, payload, skipped_at)
VALUES ($1, '{}'::jsonb, now())
ON CONFLICT (step_code) DO UPDATE
SET skipped_at   = now(),
    completed_at = NULL,
    updated_at   = now();

-- name: UpsertSystemSetting :exec
INSERT INTO system_settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value      = EXCLUDED.value,
    updated_at = now();

-- name: GetSystemSetting :one
SELECT key, value, updated_at
FROM system_settings WHERE key = $1;

-- name: CreateSetupSession :exec
INSERT INTO setup_sessions (token_hash, expires_at)
VALUES ($1, $2);

-- name: GetSetupSessionByHash :one
SELECT token_hash, expires_at, created_at
FROM setup_sessions WHERE token_hash = $1;

-- name: DeleteExpiredSetupSessions :exec
DELETE FROM setup_sessions WHERE expires_at < now();
