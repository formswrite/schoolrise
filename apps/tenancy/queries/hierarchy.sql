-- name: CreateHierarchyNode :one
INSERT INTO hierarchy_nodes (parent_id, level, code, label, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetHierarchyNodeByID :one
SELECT * FROM hierarchy_nodes
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetHierarchyNodeParent :one
SELECT * FROM hierarchy_nodes
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListHierarchyNodesByParent :many
SELECT * FROM hierarchy_nodes
WHERE COALESCE(parent_id, 0) = COALESCE($1::bigint, 0) AND deleted_at IS NULL
ORDER BY label;

-- name: ListHierarchyNodesByLevel :many
SELECT * FROM hierarchy_nodes
WHERE level = $1 AND deleted_at IS NULL
ORDER BY label;

-- name: SoftDeleteHierarchyNode :exec
UPDATE hierarchy_nodes
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: HasUndeletedChildren :one
SELECT EXISTS(
    SELECT 1 FROM hierarchy_nodes
    WHERE parent_id = $1 AND deleted_at IS NULL
) AS has_children;

-- name: InsertSelfClosure :exec
INSERT INTO hierarchy_closure (ancestor_id, descendant_id, depth)
VALUES ($1, $1, 0);

-- name: InsertClosureFromParent :exec
INSERT INTO hierarchy_closure (ancestor_id, descendant_id, depth)
SELECT ancestor_id, $2::bigint, depth + 1
FROM hierarchy_closure
WHERE descendant_id = $1::bigint;

-- name: GetDescendants :many
SELECT n.*, c.depth
FROM hierarchy_closure c
JOIN hierarchy_nodes n ON n.id = c.descendant_id
WHERE c.ancestor_id = $1 AND n.deleted_at IS NULL
ORDER BY c.depth, n.label;

-- name: GetAncestors :many
SELECT n.*, c.depth
FROM hierarchy_closure c
JOIN hierarchy_nodes n ON n.id = c.ancestor_id
WHERE c.descendant_id = $1 AND n.deleted_at IS NULL
ORDER BY c.depth DESC;

-- name: IsAncestorClosure :one
SELECT EXISTS(
    SELECT 1 FROM hierarchy_closure
    WHERE ancestor_id = $1 AND descendant_id = $2
) AS is_ancestor;

-- name: ListDescendantIDs :many
SELECT descendant_id
FROM hierarchy_closure
WHERE ancestor_id = $1;

-- name: ListAncestorIDs :many
SELECT ancestor_id
FROM hierarchy_closure
WHERE descendant_id = $1;

-- name: ListAncestorIDsForMany :many
SELECT DISTINCT ancestor_id
FROM hierarchy_closure
WHERE descendant_id = ANY($1::bigint[]);
