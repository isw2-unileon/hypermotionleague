-- s2points engine: extend player_points with the remaining match stats
-- needed to score goalkeepers/defenders and penalty events.
-- Reuses the existing player_points table (no separate player_match_stats).
-- One statement per column so the migration fails cleanly if a column
-- already exists.
ALTER TABLE player_points ADD COLUMN goals_conceded INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN pens_missed    INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN pens_saved     INT NOT NULL DEFAULT 0;
ALTER TABLE player_points ADD COLUMN saves          INT NOT NULL DEFAULT 0;
