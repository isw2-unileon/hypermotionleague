-- 007: Add the release-clause column to team_players.
-- Each owned player has a clause a rival can pay to take them without the
-- owner's consent (LaLiga Fantasy model). Nullable on purpose: rows created
-- before this feature (initial draft, won bids) stay NULL, and the clause
-- operation falls back to market_value * 2 for them.
-- ADD COLUMN IF NOT EXISTS keeps the migration idempotent (safe to re-run).
ALTER TABLE team_players ADD COLUMN IF NOT EXISTS release_clause BIGINT;
