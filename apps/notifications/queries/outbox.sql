-- name: EnqueueEmail :one
INSERT INTO notifications_outbox (kind, to_email, to_name, subject, body_html, body_text, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEmailByID :one
SELECT * FROM notifications_outbox WHERE id = $1;

-- name: ListPending :many
SELECT * FROM notifications_outbox
WHERE status = 'pending' AND scheduled_at <= now()
ORDER BY scheduled_at
LIMIT $1;

-- name: ListRecent :many
SELECT * FROM notifications_outbox
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: MarkSending :exec
UPDATE notifications_outbox
SET status = 'sending', attempts = attempts + 1, updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: MarkSent :exec
UPDATE notifications_outbox
SET status = 'sent', provider_id = $2, sent_at = now(), updated_at = now()
WHERE id = $1;

-- name: MarkFailed :exec
UPDATE notifications_outbox
SET status = $2, last_error = $3, updated_at = now()
WHERE id = $1;
