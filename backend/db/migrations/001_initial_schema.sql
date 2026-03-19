-- HyperMotion League - Initial Schema Migration
-- PostgreSQL 15+

BEGIN;

-- =============================================================================
-- EXTENSIONS
-- =============================================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- For gen_random_uuid() if needed

-- =============================================================================
-- ENUMS
-- =============================================================================
CREATE TYPE player_position AS ENUM ('GK', 'DEF', 'MID', 'FWD');
CREATE TYPE bid_status AS ENUM ('active', 'won', 'lost', 'cancelled');
CREATE TYPE league_role AS ENUM ('owner', 'admin', 'member');

-- =============================================================================
-- USERS
-- =============================================================================
CREATE TABLE users (
    id              BIGSERIAL       PRIMARY KEY,
    username        VARCHAR(50)     NOT NULL UNIQUE,
    email           VARCHAR(255)    NOT NULL UNIQUE,
    password_hash   VARCHAR(255)    NOT NULL,
    display_name    VARCHAR(100)    NOT NULL,
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_username ON users (username);

-- =============================================================================
-- LEAGUES
-- =============================================================================
CREATE TABLE leagues (
    id                  BIGSERIAL       PRIMARY KEY,
    name                VARCHAR(100)    NOT NULL,
    invite_code         VARCHAR(20)     NOT NULL UNIQUE,
    max_members         INT             NOT NULL DEFAULT 12,
    budget_per_user     INT             NOT NULL DEFAULT 100000000, -- in cents or smallest currency unit
    market_close_time   TIME            NOT NULL DEFAULT '18:00:00',
    created_by          BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leagues_invite_code ON leagues (invite_code);
CREATE INDEX idx_leagues_created_by ON leagues (created_by);

-- =============================================================================
-- LEAGUE MEMBERS (many-to-many: users <-> leagues)
-- =============================================================================
CREATE TABLE league_members (
    id          BIGSERIAL       PRIMARY KEY,
    league_id   BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id     BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        league_role     NOT NULL DEFAULT 'member',
    budget      INT             NOT NULL, -- remaining budget in this league
    joined_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_league_member UNIQUE (league_id, user_id)
);

CREATE INDEX idx_league_members_league ON league_members (league_id);
CREATE INDEX idx_league_members_user ON league_members (user_id);

-- =============================================================================
-- PLAYERS (global football player pool)
-- =============================================================================
CREATE TABLE players (
    id              BIGSERIAL       PRIMARY KEY,
    first_name      VARCHAR(50)     NOT NULL,
    last_name       VARCHAR(50)     NOT NULL,
    position        player_position NOT NULL,
    team_name       VARCHAR(100)    NOT NULL, -- real-life team (e.g., "Real Madrid")
    market_value    INT             NOT NULL DEFAULT 1000000, -- base value in cents
    is_active       BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_player_identity UNIQUE (first_name, last_name, team_name)
);

CREATE INDEX idx_players_position ON players (position);
CREATE INDEX idx_players_team ON players (team_name);
CREATE INDEX idx_players_active ON players (is_active) WHERE is_active = TRUE;

-- =============================================================================
-- TEAM PLAYERS (ownership: which user owns which player in which league)
-- =============================================================================
CREATE TABLE team_players (
    id              BIGSERIAL       PRIMARY KEY,
    league_id       BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player_id       BIGINT          NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    purchase_price  INT             NOT NULL,
    acquired_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_team_player UNIQUE (league_id, player_id),
    CONSTRAINT uq_user_player_league UNIQUE (league_id, user_id, player_id)
);

CREATE INDEX idx_team_players_league_user ON team_players (league_id, user_id);
CREATE INDEX idx_team_players_league ON team_players (league_id);
CREATE INDEX idx_team_players_player ON team_players (player_id);

-- =============================================================================
-- MATCHDAYS
-- =============================================================================
CREATE TABLE matchdays (
    id              BIGSERIAL       PRIMARY KEY,
    league_id       BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    number          INT             NOT NULL,
    name            VARCHAR(100)    NOT NULL DEFAULT '',
    start_date      TIMESTAMPTZ     NOT NULL,
    end_date        TIMESTAMPTZ     NOT NULL,
    is_current      BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_matchday_number UNIQUE (league_id, number),
    CONSTRAINT chk_matchday_dates CHECK (end_date > start_date)
);

CREATE INDEX idx_matchdays_league ON matchdays (league_id);
CREATE INDEX idx_matchdays_current ON matchdays (league_id, is_current) WHERE is_current = TRUE;
CREATE INDEX idx_matchdays_dates ON matchdays (start_date, end_date);

-- =============================================================================
-- PLAYER POINTS (points per player per matchday)
-- =============================================================================
CREATE TABLE player_points (
    id              BIGSERIAL       PRIMARY KEY,
    player_id       BIGINT          NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    matchday_id     BIGINT          NOT NULL REFERENCES matchdays(id) ON DELETE CASCADE,
    points          INT             NOT NULL DEFAULT 0,
    goals           INT             NOT NULL DEFAULT 0,
    assists         INT             NOT NULL DEFAULT 0,
    minutes_played  INT             NOT NULL DEFAULT 0,
    yellow_cards    INT             NOT NULL DEFAULT 0,
    red_cards       INT             NOT NULL DEFAULT 0,
    clean_sheet     BOOLEAN         NOT NULL DEFAULT FALSE,
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_player_matchday UNIQUE (player_id, matchday_id)
);

CREATE INDEX idx_player_points_matchday ON player_points (matchday_id);
CREATE INDEX idx_player_points_player ON player_points (player_id);
CREATE INDEX idx_player_points_player_matchday ON player_points (player_id, matchday_id);

-- =============================================================================
-- LINEUPS (user lineup per matchday per league)
-- =============================================================================
CREATE TABLE lineups (
    id              BIGSERIAL       PRIMARY KEY,
    league_id       BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    matchday_id     BIGINT          NOT NULL REFERENCES matchdays(id) ON DELETE CASCADE,
    total_points    INT             NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_lineup UNIQUE (league_id, user_id, matchday_id)
);

CREATE INDEX idx_lineups_league_matchday ON lineups (league_id, matchday_id);
CREATE INDEX idx_lineups_user ON lineups (user_id);
CREATE INDEX idx_lineups_matchday ON lineups (matchday_id);

-- =============================================================================
-- LINEUP PLAYERS (players in a lineup with position assignment)
-- =============================================================================
CREATE TABLE lineup_players (
    id              BIGSERIAL           PRIMARY KEY,
    lineup_id       BIGINT              NOT NULL REFERENCES lineups(id) ON DELETE CASCADE,
    player_id       BIGINT              NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    position        player_position     NOT NULL,
    is_starter      BOOLEAN             NOT NULL DEFAULT TRUE,
    points          INT                 NOT NULL DEFAULT 0,

    CONSTRAINT uq_lineup_player UNIQUE (lineup_id, player_id),
    CONSTRAINT uq_lineup_position UNIQUE (lineup_id, position, is_starter) -- optional: enforce position uniqueness
);

CREATE INDEX idx_lineup_players_lineup ON lineup_players (lineup_id);
CREATE INDEX idx_lineup_players_player ON lineup_players (player_id);

-- =============================================================================
-- MARKET LISTINGS (players available in a league's market)
-- =============================================================================
CREATE TABLE market_listings (
    id              BIGSERIAL       PRIMARY KEY,
    league_id       BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    player_id       BIGINT          NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    base_price      INT             NOT NULL,
    seller_id       BIGINT          REFERENCES users(id) ON DELETE SET NULL, -- NULL if from global pool
    status          VARCHAR(20)     NOT NULL DEFAULT 'active', -- active, sold, cancelled
    listed_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ     NOT NULL,

    CONSTRAINT uq_market_listing UNIQUE (league_id, player_id, status) -- prevent duplicate active listings
);

CREATE INDEX idx_market_listings_league ON market_listings (league_id);
CREATE INDEX idx_market_listings_status ON market_listings (league_id, status);
CREATE INDEX idx_market_listings_expires ON market_listings (expires_at) WHERE status = 'active';

-- =============================================================================
-- BIDS (user bids on market listings)
-- =============================================================================
CREATE TABLE bids (
    id              BIGSERIAL       PRIMARY KEY,
    listing_id      BIGINT          NOT NULL REFERENCES market_listings(id) ON DELETE CASCADE,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount          INT             NOT NULL,
    status          bid_status      NOT NULL DEFAULT 'active',
    placed_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_bid_amount_positive CHECK (amount > 0)
);

CREATE INDEX idx_bids_listing ON bids (listing_id);
CREATE INDEX idx_bids_user ON bids (user_id);
CREATE INDEX idx_bids_user_active ON bids (user_id, status) WHERE status = 'active';
CREATE INDEX idx_bids_listing_amount ON bids (listing_id, amount DESC);

-- =============================================================================
-- HELPER FUNCTIONS
-- =============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Auto-update updated_at triggers
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_leagues_updated_at
    BEFORE UPDATE ON leagues
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_players_updated_at
    BEFORE UPDATE ON players
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_lineups_updated_at
    BEFORE UPDATE ON lineups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMIT;
