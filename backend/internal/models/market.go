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

// MarketStatus represents the current state of a league market.
type MarketStatus struct {
	LeagueID       int64     `json:"league_id"`
	IsOpen         bool      `json:"is_open"`
	ClosesAt       time.Time `json:"closes_at"`
	ActiveListings int       `json:"active_listings"`
	YourActiveBids int       `json:"your_active_bids"`
	MaxBidsPerUser int       `json:"max_bids_per_user"`
}
