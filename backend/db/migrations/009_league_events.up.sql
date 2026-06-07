-- 009: Activity feed for leagues.
CREATE TABLE league_events (
    id          BIGSERIAL       PRIMARY KEY,
    league_id   BIGINT          NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    user_id     BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type  TEXT            NOT NULL,
    player_name TEXT,
    amount      INT,
    details     TEXT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_league_events_league ON league_events (league_id, created_at DESC);
