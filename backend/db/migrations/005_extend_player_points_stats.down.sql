-- Revert 005: drop the Sprint 2 points-engine stat columns from player_points.
ALTER TABLE player_points DROP COLUMN IF EXISTS goals_conceded;
ALTER TABLE player_points DROP COLUMN IF EXISTS pens_missed;
ALTER TABLE player_points DROP COLUMN IF EXISTS pens_saved;
ALTER TABLE player_points DROP COLUMN IF EXISTS saves;
