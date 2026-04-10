-- =============================================================================
-- TEAMS (real-world football teams from API-Football)
-- =============================================================================
CREATE TABLE IF NOT EXISTS teams (
    id              BIGINT          PRIMARY KEY,  -- API-Football team ID
    name            VARCHAR(100)    NOT NULL,
    code            VARCHAR(10),
    logo_url        TEXT,
    country         VARCHAR(50),
    founded         INT,
    venue_name      VARCHAR(200),
    venue_city      VARCHAR(100),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- PLAYERS — add API-Football fields
-- =============================================================================
-- External ID for matching with API-Football
ALTER TABLE players ADD COLUMN IF NOT EXISTS api_football_id BIGINT UNIQUE;

-- Photo URL from API-Football
ALTER TABLE players ADD COLUMN IF NOT EXISTS photo_url TEXT;

-- Link to teams table (nullable — existing rows don't have it yet)
ALTER TABLE players ADD COLUMN IF NOT EXISTS team_id BIGINT REFERENCES teams(id) ON DELETE SET NULL;

-- Age and nationality (useful metadata)
ALTER TABLE players ADD COLUMN IF NOT EXISTS age INT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS nationality VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_players_api_football_id ON players (api_football_id);
CREATE INDEX IF NOT EXISTS idx_players_team_id ON players (team_id);

-- =============================================================================
-- FIXTURES (real-world match data from API-Football)
-- =============================================================================
CREATE TABLE IF NOT EXISTS fixtures (
    id              BIGINT          PRIMARY KEY,  -- API-Football fixture ID
    league_id_api   INT             NOT NULL,     -- API-Football league ID (141 for Hypermotion)
    season          INT             NOT NULL,
    round           VARCHAR(50),                  -- e.g. "Regular Season - 1"
    home_team_id    BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    away_team_id    BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    home_goals      INT,
    away_goals      INT,
    status          VARCHAR(20)     NOT NULL DEFAULT 'NS',  -- NS, FT, 1H, HT, 2H, etc.
    match_date      TIMESTAMPTZ     NOT NULL,
    venue_name      VARCHAR(200),
    venue_city      VARCHAR(100),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fixtures_season ON fixtures (league_id_api, season);
CREATE INDEX IF NOT EXISTS idx_fixtures_round ON fixtures (round);
CREATE INDEX IF NOT EXISTS idx_fixtures_date ON fixtures (match_date);
CREATE INDEX IF NOT EXISTS idx_fixtures_home_team ON fixtures (home_team_id);
CREATE INDEX IF NOT EXISTS idx_fixtures_away_team ON fixtures (away_team_id);

CREATE TRIGGER trg_fixtures_updated_at
    BEFORE UPDATE ON fixtures
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
