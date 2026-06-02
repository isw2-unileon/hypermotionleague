package models

import "time"

// Team represents a real-world football club loaded from API-Football.
//
// This is the real-club catalog and is unrelated to fantasy ownership
// (TeamPlayer / UserTeam below). Founded is a pointer so an unknown value
// (API-Football null / 0) is stored as SQL NULL rather than a literal 0.
type Team struct {
	ID         int64     `json:"id"`
	ExternalID int64     `json:"external_id"` // API-Football team.id
	Name       string    `json:"name"`
	Code       string    `json:"code,omitempty"`
	Country    string    `json:"country,omitempty"`
	Founded    *int      `json:"founded,omitempty"`
	LogoURL    string    `json:"logo_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TeamPlayer represents a player owned by a user in a specific league.
type TeamPlayer struct {
	ID            int64     `json:"id"`
	LeagueID      int64     `json:"league_id"`
	UserID        int64     `json:"user_id"`
	PlayerID      int64     `json:"player_id"`
	PurchasePrice int       `json:"purchase_price"`
	AcquiredAt    time.Time `json:"acquired_at"`
}

// TeamPlayerWithDetails includes player and owner info.
type TeamPlayerWithDetails struct {
	TeamPlayer
	Player    Player `json:"player"`
	OwnerName string `json:"owner_name"`
}

// UserTeam represents a user's squad in a league.
type UserTeam struct {
	LeagueID    int64                   `json:"league_id"`
	UserID      int64                   `json:"user_id"`
	Username    string                  `json:"username"`
	DisplayName string                  `json:"display_name"`
	Budget      int                     `json:"budget"`
	Players     []TeamPlayerWithDetails `json:"players"`
	TotalValue  int                     `json:"total_value"`
}
