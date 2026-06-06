-- 007: Allow lineups without a matchday (default/pre-season lineup).
-- matchday_id NULL = the user's default formation for that league.
ALTER TABLE lineups ALTER COLUMN matchday_id DROP NOT NULL;

-- Update the unique constraint: allow one default lineup (NULL matchday) per user per league.
-- The original uq_lineup (league_id, user_id, matchday_id) doesn't prevent duplicates
-- when matchday_id is NULL (SQL NULLs are never equal). Add a partial unique index.
DROP INDEX IF EXISTS idx_lineup_default;
CREATE UNIQUE INDEX idx_lineup_default ON lineups (league_id, user_id)
  WHERE matchday_id IS NULL;
