ALTER TABLE users
    ADD COLUMN IF NOT EXISTS show_in_leaderboard BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS leaderboard_alias   VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_leaderboard_alias
    ON users (leaderboard_alias)
    WHERE leaderboard_alias IS NOT NULL AND leaderboard_alias <> '';

CREATE INDEX IF NOT EXISTS idx_users_leaderboard_consent
    ON users (show_in_leaderboard)
    WHERE show_in_leaderboard = true;
