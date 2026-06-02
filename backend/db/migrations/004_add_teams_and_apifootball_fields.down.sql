-- Revert 004_add_teams_and_apifootball_fields.
-- Reverse order of the up migration.

-- =============================================================================
-- PLAYERS (remove API-Football fields)
-- =============================================================================
DROP INDEX IF EXISTS idx_players_team_id;

ALTER TABLE players DROP CONSTRAINT IF EXISTS uq_players_external_id;

ALTER TABLE players
    DROP COLUMN IF EXISTS age,
    DROP COLUMN IF EXISTS photo_url,
    DROP COLUMN IF EXISTS jersey_number,
    DROP COLUMN IF EXISTS team_id,
    DROP COLUMN IF EXISTS external_id;

-- Restore the original identity guard from 001_initial_schema.up.sql.
-- WARNING: this fails if synced data left duplicate (first_name, last_name,
-- team_name) rows; de-duplicate before rolling back if so.
ALTER TABLE players ADD CONSTRAINT uq_player_identity UNIQUE (first_name, last_name, team_name);

-- =============================================================================
-- TEAMS
-- =============================================================================
DROP TRIGGER IF EXISTS trg_teams_updated_at ON teams;
DROP TABLE IF EXISTS teams;
