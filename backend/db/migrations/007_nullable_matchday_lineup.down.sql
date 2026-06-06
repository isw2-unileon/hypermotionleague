DROP INDEX IF EXISTS idx_lineup_default;
DELETE FROM lineups WHERE matchday_id IS NULL;
ALTER TABLE lineups ALTER COLUMN matchday_id SET NOT NULL;
