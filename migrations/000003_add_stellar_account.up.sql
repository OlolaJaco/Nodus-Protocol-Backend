ALTER TABLE users
    ADD COLUMN IF NOT EXISTS stellar_account_id VARCHAR(56) UNIQUE;

CREATE INDEX IF NOT EXISTS idx_users_stellar_account_id ON users(stellar_account_id)
    WHERE stellar_account_id IS NOT NULL;
