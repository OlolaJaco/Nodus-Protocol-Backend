ALTER TABLE users ADD COLUMN IF NOT EXISTS stellar_address VARCHAR(60);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_stellar_address ON users (stellar_address) WHERE stellar_address IS NOT NULL AND stellar_address <> '';
