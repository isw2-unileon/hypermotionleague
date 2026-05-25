-- Drop the overly restrictive uq_lineup_position constraint.
-- It only allowed one starter per position per lineup, which blocks
-- any real formation (e.g. a 4-4-2 needs 4 DEF starters).
-- The uq_lineup_player UNIQUE (lineup_id, player_id) constraint already
-- enforces the real invariant (no duplicate players within a lineup).
ALTER TABLE lineup_players DROP CONSTRAINT uq_lineup_position;
