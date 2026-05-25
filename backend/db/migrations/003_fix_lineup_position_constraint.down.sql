-- Restore the original uq_lineup_position constraint from
-- 001_initial_schema.up.sql.
ALTER TABLE lineup_players
  ADD CONSTRAINT uq_lineup_position UNIQUE (lineup_id, position, is_starter);
