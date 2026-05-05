\c auth

CREATE TABLE users (
    id                   BIGSERIAL PRIMARY KEY,
    email                TEXT NOT NULL,
    password_hash        TEXT NOT NULL,
    full_name            TEXT NOT NULL,
    role                 TEXT NOT NULL,
    must_change_password BOOLEAN NOT NULL DEFAULT false,
    locked_at            TIMESTAMPTZ,
    failed_attempts      INT NOT NULL DEFAULT 0,
    last_login_at        TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_email_lower
    ON users (LOWER(email))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_users_role ON users (role) WHERE deleted_at IS NULL;
