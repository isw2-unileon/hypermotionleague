-- Add a real-world football clubs table and API-Football fields to players.
-- These back the `sync-players` command, which loads La Liga Hypermotion
-- (Segunda División, league 141) teams and squads from API-Football.
--
-- NOTE: this `teams` table holds REAL clubs and is unrelated to the fantasy
-- ownership table `team_players` (which user owns which player in a league).

-- =============================================================================
-- TEAMS (real football clubs from API-Football)
-- =============================================================================
CREATE TABLE IF NOT EXISTS teams (
    id          BIGSERIAL    PRIMARY KEY,
    external_id BIGINT       NOT NULL,                -- API-Football team.id (upsert key)
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(10),                          -- e.g. "RMA" (nullable)
    country     VARCHAR(100),
    founded     INT,
    logo_url    TEXT,                                 -- media.api-sports.io URL only, never the image
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_teams_external_id UNIQUE (external_id)
);

-- Reuse the shared updated_at trigger function defined in 001_initial_schema.
CREATE OR REPLACE TRIGGER trg_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- PLAYERS (add API-Football fields)
-- =============================================================================

-- external_id (API-Football player.id) is now the canonical player identity and
-- upsert key, so the old name+team identity guard is dropped. It would also
-- reject the rare case of two distinct players whose combined name splits to the
-- same first/last within the same club.
ALTER TABLE players DROP CONSTRAINT IF EXISTS uq_player_identity;

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS external_id   BIGINT,                                   -- API-Football player.id (upsert key)
    ADD COLUMN IF NOT EXISTS team_id       BIGINT REFERENCES teams(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS jersey_number INT,
    ADD COLUMN IF NOT EXISTS photo_url     TEXT,                                     -- media.api-sports.io URL only
    ADD COLUMN IF NOT EXISTS age           INT;

-- UNIQUE on external_id is the upsert key. Postgres allows multiple NULLs, so
-- pre-existing players (external_id NULL) are unaffected. Guarded so the
-- migration can be re-run safely.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_players_external_id'
    ) THEN
        ALTER TABLE players ADD CONSTRAINT uq_players_external_id UNIQUE (external_id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_players_team_id ON players (team_id);
