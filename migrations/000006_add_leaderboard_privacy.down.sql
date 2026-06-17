ALTER TABLE users
    DROP COLUMN IF EXISTS show_in_leaderboard,
    DROP COLUMN IF EXISTS leaderboard_alias;
