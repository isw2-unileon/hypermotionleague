-- 008: Fix the duplicate-bid constraint so it only applies to ACTIVE bids.
--
-- 006 added uq_bid_listing_user UNIQUE (listing_id, user_id) with no status
-- filter, so a user kept a single bids row per listing forever. Cancelling a
-- bid only sets status='cancelled' (the row stays), so re-bidding on the same
-- listing hit the constraint and PlaceBidTx returned a 500.
--
-- Replace the full UNIQUE with a partial unique index scoped to active bids:
-- a user still cannot hold two ACTIVE bids on the same listing, but cancelled/
-- won/lost rows no longer block a new bid, preserving bid history.
ALTER TABLE bids DROP CONSTRAINT IF EXISTS uq_bid_listing_user;
CREATE UNIQUE INDEX IF NOT EXISTS uq_bid_listing_user_active
    ON bids (listing_id, user_id)
    WHERE status = 'active';
