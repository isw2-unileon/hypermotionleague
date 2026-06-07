package models

import "time"

// BidStatus represents the state of a bid.
type BidStatus string

const (
	BidActive    BidStatus = "active"
	BidWon       BidStatus = "won"
	BidLost      BidStatus = "lost"
	BidCancelled BidStatus = "cancelled"
)

// MarketListing represents a player available for purchase in a league.
type MarketListing struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"league_id"`
	PlayerID  int64     `json:"player_id"`
	BasePrice int       `json:"base_price"`
	SellerID  *int64    `json:"seller_id,omitempty"` // nil if from global pool
	Status    string    `json:"status"`
	ListedAt  time.Time `json:"listed_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MarketListingWithDetails includes player and bid info.
type MarketListingWithDetails struct {
	MarketListing
	Player     Player  `json:"player"`
	SellerName *string `json:"seller_name,omitempty"`
	HighestBid *int    `json:"highest_bid,omitempty"`
	BidCount   int     `json:"bid_count"`
}

// Bid represents a user's offer on a market listing.
type Bid struct {
	ID        int64     `json:"id"`
	ListingID int64     `json:"listing_id"`
	UserID    int64     `json:"user_id"`
	Amount    int       `json:"amount"`
	Status    BidStatus `json:"status"`
	PlacedAt  time.Time `json:"placed_at"`
}

// BidWithDetails includes listing and player info.
type BidWithDetails struct {
	Bid
	Listing MarketListing `json:"listing"`
	Player  Player        `json:"player"`
}

// PlaceBidRequest is the payload for placing a bid.
type PlaceBidRequest struct {
	ListingID int64 `json:"listing_id" binding:"required"`
	Amount    int   `json:"amount"     binding:"required,min=1"`
}

// ActivityEvent represents a single event in a league activity feed.
type ActivityEvent struct {
	ID         int64     `json:"id"`
	LeagueID   int64     `json:"league_id"`
	UserID     int64     `json:"user_id"`
	EventType  string    `json:"event_type"` // join, leave, bid, bid_cancel, bid_raise, transfer, top_scorer
	Actor      string    `json:"actor"`
	PlayerName string    `json:"player_name,omitempty"`
	Amount     int       `json:"amount,omitempty"`
	Details    string    `json:"details,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

// ClauseResult is the outcome of a successful release-clause purchase, returned
// to the buyer. Amounts are in the same unit as budget/purchase_price.
type ClauseResult struct {
	LeagueID         int64  `json:"league_id"`
	PlayerID         int64  `json:"player_id"`
	PlayerName       string `json:"player_name"`
	PreviousOwnerID  int64  `json:"previous_owner_id"`
	NewOwnerID       int64  `json:"new_owner_id"`       // the buyer
	AmountPaid       int    `json:"amount_paid"`        // clause charged
	NewReleaseClause int    `json:"new_release_clause"` // recomputed for the new owner
	BuyerBudget      int    `json:"buyer_budget"`       // buyer budget after paying
	SellerBudget     int    `json:"seller_budget"`      // seller budget after receiving
}

// MarketStatus represents the current state of a league market.
//
// IsOpen / NextChangeAt / Reason describe the trading window (see package
// market): is_open requires both the 19:00–00:00 Europe/Madrid window and no
// matchday in play. NextChangeAt is the absolute instant of the next open<->
// closed flip (the frontend renders the countdown from it). Reason is one of
// "open" | "outside_window" | "active_matchday".
type MarketStatus struct {
	LeagueID       int64     `json:"league_id"`
	IsOpen         bool      `json:"is_open"`
	NextChangeAt   time.Time `json:"next_change_at"`
	Reason         string    `json:"reason"`
	ActiveListings int       `json:"active_listings"`
	YourActiveBids int       `json:"your_active_bids"`
	MaxBidsPerUser int       `json:"max_bids_per_user"`
}
