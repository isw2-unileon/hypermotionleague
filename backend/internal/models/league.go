package models

import "time"

// LeagueRole defines the role of a member within a league.
type LeagueRole string

const (
	RoleOwner  LeagueRole = "owner"
	RoleAdmin  LeagueRole = "admin"
	RoleMember LeagueRole = "member"
)

// League represents a fantasy football league.
type League struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	InviteCode      string    `json:"invite_code"`
	MaxMembers      int       `json:"max_members"`
	BudgetPerUser   int       `json:"budget_per_user"`
	MarketCloseTime string    `json:"market_close_time"`
	CreatedBy       int64     `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// LeagueMember represents a user's membership in a league.
type LeagueMember struct {
	ID       int64      `json:"id"`
	LeagueID int64      `json:"league_id"`
	UserID   int64      `json:"user_id"`
	Role     LeagueRole `json:"role"`
	Budget   int        `json:"budget"`
	JoinedAt time.Time  `json:"joined_at"`
}

// LeagueWithMembers includes member details for display.
type LeagueWithMembers struct {
	League
	Members []LeagueMember `json:"members"`
}

// CreateLeagueRequest is the payload for creating a league.
type CreateLeagueRequest struct {
	Name            string `json:"name"           binding:"required,min=1,max=100"`
	MaxMembers      int    `json:"max_members"    binding:"omitempty,min=2,max=20"`
	BudgetPerUser   int    `json:"budget_per_user" binding:"omitempty,min=1000000"`
	MarketCloseTime string `json:"market_close_time" binding:"omitempty"`
}

// JoinLeagueRequest is the payload for joining a league.
type JoinLeagueRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// LoginRequest is the payload for logging in.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the response returned after a successful login.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
