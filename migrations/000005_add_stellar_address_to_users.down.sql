DROP INDEX IF EXISTS idx_users_stellar_address;
ALTER TABLE users DROP COLUMN IF EXISTS stellar_address;
