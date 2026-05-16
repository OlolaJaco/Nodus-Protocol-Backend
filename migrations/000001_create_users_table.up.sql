-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email            VARCHAR(255) NOT NULL UNIQUE,
    password_hash    VARCHAR(255) NOT NULL,
    first_name       VARCHAR(100) NOT NULL DEFAULT '',
    last_name        VARCHAR(100) NOT NULL DEFAULT '',
    avatar_url       VARCHAR(512),
    role             VARCHAR(20)  NOT NULL DEFAULT 'user',
    is_email_verified BOOLEAN     NOT NULL DEFAULT false,
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    last_login_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_users_email        ON users(email);
CREATE INDEX idx_users_deleted_at   ON users(deleted_at);
CREATE INDEX idx_users_is_active    ON users(is_active);
