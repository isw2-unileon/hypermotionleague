-- 006: Prevent a user from placing duplicate bids on the same listing.
-- Without this, a user can inflate the 5-bid cap and complicate budget calc.
--
-- Guarded so the migration is idempotent: re-running it (e.g. after this file
-- was renumbered 004 -> 005 -> 006 and the tracking saw it as new) is a no-op
-- when the constraint already exists, rather than erroring out. Same pattern as
-- uq_players_external_id in 004_add_teams_and_apifootball_fields.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_bid_listing_user'
    ) THEN
        ALTER TABLE bids ADD CONSTRAINT uq_bid_listing_user UNIQUE (listing_id, user_id);
    END IF;
END $$;
