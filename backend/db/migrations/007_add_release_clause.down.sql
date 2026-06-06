-- Revert 007: drop the release-clause column from team_players.
ALTER TABLE team_players DROP COLUMN IF EXISTS release_clause;
