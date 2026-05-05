\c tenancy

CREATE TABLE hierarchy_levels (
    code              TEXT PRIMARY KEY,
    label             TEXT NOT NULL,
    parent_level_code TEXT REFERENCES hierarchy_levels(code) ON DELETE RESTRICT,
    depth             INT NOT NULL,
    sort_order        INT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_hierarchy_levels_parent ON hierarchy_levels (parent_level_code);
CREATE INDEX idx_hierarchy_levels_depth  ON hierarchy_levels (depth);
