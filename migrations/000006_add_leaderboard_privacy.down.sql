DROP INDEX IF EXISTS idx_users_leaderboard_alias;
DROP INDEX IF EXISTS idx_users_leaderboard_consent;

ALTER TABLE users
    DROP COLUMN IF EXISTS show_in_leaderboard,
    DROP COLUMN IF EXISTS leaderboard_alias;
