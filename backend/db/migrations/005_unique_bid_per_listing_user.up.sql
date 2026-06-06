-- 004: Prevent a user from placing duplicate bids on the same listing.
-- Without this, a user can inflate the 5-bid cap and complicate budget calc.
ALTER TABLE bids ADD CONSTRAINT uq_bid_listing_user UNIQUE (listing_id, user_id);
