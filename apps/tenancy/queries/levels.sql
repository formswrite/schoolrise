-- name: ListHierarchyLevels :many
SELECT code, label, parent_level_code, depth, sort_order, created_at, updated_at
FROM hierarchy_levels
ORDER BY depth, sort_order, code;

-- name: GetHierarchyLevel :one
SELECT code, label, parent_level_code, depth, sort_order, created_at, updated_at
FROM hierarchy_levels
WHERE code = $1;

-- name: UpsertHierarchyLevel :exec
INSERT INTO hierarchy_levels (code, label, parent_level_code, depth, sort_order)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (code) DO UPDATE
SET label             = EXCLUDED.label,
    parent_level_code = EXCLUDED.parent_level_code,
    depth             = EXCLUDED.depth,
    sort_order        = EXCLUDED.sort_order,
    updated_at        = now();

-- name: DeleteAllHierarchyLevels :exec
DELETE FROM hierarchy_levels;
