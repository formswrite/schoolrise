\c auth

CREATE TABLE role_assignments (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL,
    scope_node_id BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_role_assignments_uniq
    ON role_assignments (user_id, role, COALESCE(scope_node_id, 0));

CREATE INDEX idx_role_assignments_user ON role_assignments (user_id);
CREATE INDEX idx_role_assignments_scope ON role_assignments (scope_node_id);
