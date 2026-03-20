-- HyperMotion League - Initial Schema (DOWN)
-- Drops all objects in reverse dependency order

-- =============================================================================
-- TRIGGERS
-- =============================================================================
DROP TRIGGER IF EXISTS trg_lineups_updated_at ON lineups;
DROP TRIGGER IF EXISTS trg_players_updated_at ON players;
DROP TRIGGER IF EXISTS trg_leagues_updated_at ON leagues;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

-- =============================================================================
-- FUNCTIONS
-- =============================================================================
DROP FUNCTION IF EXISTS update_updated_at_column();

-- =============================================================================
-- TABLES (reverse dependency order)
-- =============================================================================
DROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS market_listings;
DROP TABLE IF EXISTS lineup_players;
DROP TABLE IF EXISTS lineups;
DROP TABLE IF EXISTS player_points;
DROP TABLE IF EXISTS matchdays;
DROP TABLE IF EXISTS team_players;
DROP TABLE IF EXISTS players;
DROP TABLE IF EXISTS league_members;
DROP TABLE IF EXISTS leagues;
DROP TABLE IF EXISTS users;

-- =============================================================================
-- ENUMS
-- =============================================================================
DROP TYPE IF EXISTS league_role;
DROP TYPE IF EXISTS bid_status;
DROP TYPE IF EXISTS player_position;
