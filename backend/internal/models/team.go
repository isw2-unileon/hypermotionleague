package models

import "time"

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
