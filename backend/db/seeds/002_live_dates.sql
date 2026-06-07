-- Live-dates seed: makes the app show something active when run today.
-- Safe to re-run (idempotent UPDATE + upsert).
-- Do NOT apply to the shared/production DB.

BEGIN;

-- ── Matchdays: reset current flag, then set live dates ────────────────────────

-- Matchdays 1 and 2 stay in the past (player_points demo data references them).
-- Matchday 3 becomes the active one, centered on today.
-- Matchday 4 becomes the next upcoming one.

-- Matchdays are global now, so reset the current flag across all of them.
UPDATE matchdays SET is_current = FALSE;

-- Past matchday with points data (keep close to now so demo feels recent)
UPDATE matchdays SET
    start_date = (NOW() - INTERVAL '14 days'),
    end_date   = (NOW() - INTERVAL '8 days'),
    is_current = FALSE
WHERE number = 1;

UPDATE matchdays SET
    start_date = (NOW() - INTERVAL '7 days'),
    end_date   = (NOW() - INTERVAL '1 day'),
    is_current = FALSE
WHERE number = 2;

-- Current matchday: open window around today (deadline ~5 days from now)
UPDATE matchdays SET
    start_date = (NOW() - INTERVAL '1 day'),
    end_date   = (NOW() + INTERVAL '5 days'),
    is_current = TRUE
WHERE number = 3;

-- Future matchday
UPDATE matchdays SET
    start_date = (NOW() + INTERVAL '7 days'),
    end_date   = (NOW() + INTERVAL '14 days'),
    is_current = FALSE
WHERE number = 4;

-- ── Market listings: ensure at least one open listing has a future expiry ─────

-- Haaland listing — bump expiry to 7 days from now so the market is open.
-- ON CONFLICT handles the case where the listing was already resolved/re-inserted.
INSERT INTO market_listings (league_id, player_id, base_price, seller_id, status, listed_at, expires_at)
VALUES (1, 19, 38000000, NULL, 'active', NOW(), NOW() + INTERVAL '7 days')
ON CONFLICT (league_id, player_id, status) DO UPDATE SET
    listed_at  = NOW(),
    expires_at = NOW() + INTERVAL '7 days';

-- Saka listing — also keep open
INSERT INTO market_listings (league_id, player_id, base_price, seller_id, status, listed_at, expires_at)
VALUES (1, 23, 25000000, NULL, 'active', NOW(), NOW() + INTERVAL '7 days')
ON CONFLICT (league_id, player_id, status) DO UPDATE SET
    listed_at  = NOW(),
    expires_at = NOW() + INTERVAL '7 days';

COMMIT;
