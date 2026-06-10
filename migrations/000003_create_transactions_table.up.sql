CREATE TABLE IF NOT EXISTS transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    engine_id   VARCHAR(36) NOT NULL UNIQUE,
    sender      VARCHAR(60) NOT NULL,
    recipient   VARCHAR(60) NOT NULL,
    amount      BIGINT      NOT NULL CHECK (amount > 0),
    token       VARCHAR(12) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    tx_hash     VARCHAR(256),
    fee_stroops BIGINT      NOT NULL DEFAULT 0,
    urgency     VARCHAR(20) NOT NULL DEFAULT 'standard',
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id  ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status   ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_engine_id ON transactions(engine_id);
