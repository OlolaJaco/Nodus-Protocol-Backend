ALTER TABLE users
    ADD COLUMN IF NOT EXISTS show_in_leaderboard BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS leaderboard_alias   VARCHAR(32);
