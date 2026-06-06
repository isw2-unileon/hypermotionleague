-- s2points engine: extend player_points with the remaining match stats
-- needed to score goalkeepers/defenders and penalty events.
-- Reuses the existing player_points table (no separate player_match_stats).
-- ADD COLUMN IF NOT EXISTS so the migration is idempotent: re-running it (e.g.
-- after this file was renumbered 004 -> 005 and the tracking saw it as new)
-- is a no-op when the columns already exist, rather than erroring out.
ALTER TABLE player_points ADD COLUMN IF NOT EXISTS goals_conceded INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN IF NOT EXISTS pens_missed    INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN IF NOT EXISTS pens_saved     INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN IF NOT EXISTS saves          INT NOT NULL DEFAULT 0;
