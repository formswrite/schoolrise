\c tenancy

CREATE TABLE hierarchy_nodes (
    id          BIGSERIAL PRIMARY KEY,
    parent_id   BIGINT REFERENCES hierarchy_nodes(id) ON DELETE RESTRICT,
    level       TEXT NOT NULL,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    metadata    JSONB NOT NULL DEFAULT '{}'::jsonb,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_hierarchy_nodes_parent_level_code
    ON hierarchy_nodes (COALESCE(parent_id, 0), level, code)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_hierarchy_nodes_parent
    ON hierarchy_nodes (parent_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_hierarchy_nodes_level
    ON hierarchy_nodes (level)
    WHERE deleted_at IS NULL;

CREATE TABLE hierarchy_closure (
    ancestor_id   BIGINT NOT NULL REFERENCES hierarchy_nodes(id) ON DELETE CASCADE,
    descendant_id BIGINT NOT NULL REFERENCES hierarchy_nodes(id) ON DELETE CASCADE,
    depth         INT NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX idx_closure_descendant ON hierarchy_closure (descendant_id);
CREATE INDEX idx_closure_depth ON hierarchy_closure (depth);
