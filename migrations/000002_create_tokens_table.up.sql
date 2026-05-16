CREATE TABLE IF NOT EXISTS tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hash       VARCHAR(255) NOT NULL UNIQUE,
    type       VARCHAR(30) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tokens_user_id    ON tokens(user_id);
CREATE INDEX idx_tokens_type       ON tokens(type);
CREATE INDEX idx_tokens_deleted_at ON tokens(deleted_at);
CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);
