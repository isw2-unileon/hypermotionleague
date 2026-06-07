-- Revert 009: restore per-league matchdays.
--
-- NOTE: the original league_id was BIGINT NOT NULL REFERENCES leagues(id). It
-- cannot be restored NOT NULL without backfilling a league per matchday, so this
-- down migration re-adds it as NULLABLE (the game tables are empty in practice).
-- Backfill league_id and re-apply NOT NULL by hand if you have real rows.

DROP INDEX IF EXISTS idx_matchdays_current;
ALTER TABLE matchdays DROP CONSTRAINT IF EXISTS uq_matchday_number; -- was (number)

ALTER TABLE matchdays
    ADD COLUMN IF NOT EXISTS league_id BIGINT REFERENCES leagues(id) ON DELETE CASCADE;

ALTER TABLE matchdays ADD CONSTRAINT uq_matchday_number UNIQUE (league_id, number);

CREATE INDEX IF NOT EXISTS idx_matchdays_league ON matchdays (league_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_matchdays_current ON matchdays (league_id, is_current)
  WHERE is_current = TRUE;
