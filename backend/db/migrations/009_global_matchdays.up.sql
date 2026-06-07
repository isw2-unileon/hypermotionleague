-- 009: Make matchdays GLOBAL (app-wide) instead of per-league.
--
-- A matchday is now one loaded round of the real competition (La Liga 2); its
-- `number` is the app's OWN sequential counter (1, 2, 3...), independent of the
-- real Segunda round. player_points and lineups already key off matchdays.id
-- (player_points has no league_id; lineups carry their own league_id), so making
-- matchdays global is just removing matchdays.league_id and globalizing its
-- constraints — nothing else needs to move.
--
-- Game tables are ~empty in production (0 matchdays / lineups / player_points),
-- so dropping and recreating constraints is clean. Re-runnable.

-- Per-league indexes/constraint go away.
DROP INDEX IF EXISTS idx_matchdays_league;
DROP INDEX IF EXISTS idx_matchdays_current;
ALTER TABLE matchdays DROP CONSTRAINT IF EXISTS uq_matchday_number; -- was (league_id, number)

-- Drop the per-league scope. This also removes the matchdays -> leagues FK.
ALTER TABLE matchdays DROP COLUMN IF EXISTS league_id;

-- `number` is now the app's globally-unique sequential matchday counter.
ALTER TABLE matchdays ADD CONSTRAINT uq_matchday_number UNIQUE (number);

-- At most one current matchday app-wide (partial unique index over the single
-- TRUE value allows only one row with is_current = TRUE).
CREATE UNIQUE INDEX IF NOT EXISTS idx_matchdays_current ON matchdays (is_current)
  WHERE is_current = TRUE;

-- chk_matchday_dates and idx_matchdays_dates are unchanged (kept as-is).
