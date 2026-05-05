-- name: CreateRoleAssignment :exec
INSERT INTO role_assignments (user_id, role, scope_node_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, role, COALESCE(scope_node_id, 0)) DO NOTHING;

-- name: DeleteRoleAssignment :exec
DELETE FROM role_assignments WHERE id = $1;

-- name: ListRoleAssignmentsForUser :many
SELECT id, user_id, role, scope_node_id, created_at
FROM role_assignments
WHERE user_id = $1
ORDER BY created_at;

-- name: ListAllRoleAssignments :many
SELECT id, user_id, role, scope_node_id, created_at
FROM role_assignments
ORDER BY user_id, created_at;
