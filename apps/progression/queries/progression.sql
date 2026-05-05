-- name: UpsertSnapshot :exec
INSERT INTO progression_snapshots (scope_node_id, period_id, campaign_id, band_code, band_ordinal, student_count)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (scope_node_id, period_id, campaign_id, band_code) DO UPDATE
SET student_count = EXCLUDED.student_count,
    band_ordinal = EXCLUDED.band_ordinal,
    snapshot_at = now();

-- name: ListSnapshotsForScope :many
SELECT * FROM progression_snapshots
WHERE scope_node_id = $1 AND period_id = $2 AND campaign_id = $3
ORDER BY band_ordinal;

-- name: ListSnapshotsForCampaign :many
SELECT * FROM progression_snapshots
WHERE campaign_id = $1 AND period_id = $2
ORDER BY scope_node_id, band_ordinal;

-- name: DeleteSnapshotsForCampaign :exec
DELETE FROM progression_snapshots
WHERE campaign_id = $1 AND period_id = $2;

-- name: CreateRefreshLog :one
INSERT INTO progression_refresh_log (campaign_id)
VALUES ($1)
RETURNING *;

-- name: CompleteRefreshLog :exec
UPDATE progression_refresh_log
SET completed_at = now(), rows_written = $2, error = $3
WHERE id = $1;

-- name: ListRecentRefreshes :many
SELECT * FROM progression_refresh_log
WHERE campaign_id = $1
ORDER BY triggered_at DESC
LIMIT $2;
