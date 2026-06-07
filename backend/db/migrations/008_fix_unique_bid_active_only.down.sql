-- Revert 008: drop the partial unique index and restore the original full
-- UNIQUE (listing_id, user_id) constraint added in 006.
DROP INDEX IF EXISTS uq_bid_listing_user_active;
ALTER TABLE bids ADD CONSTRAINT uq_bid_listing_user UNIQUE (listing_id, user_id);
