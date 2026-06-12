DROP INDEX IF EXISTS idx_users_stellar_account_id;
ALTER TABLE users DROP COLUMN IF EXISTS stellar_account_id;
