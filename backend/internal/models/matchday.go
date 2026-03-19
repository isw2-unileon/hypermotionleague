package models

import "time"

// Matchday represents a game week within a league.
type Matchday struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"league_id"`
	Number    int       `json:"number"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsCurrent bool      `json:"is_current"`
	CreatedAt time.Time `json:"created_at"`
}

// Lineup represents a user's selected players for a matchday.
type Lineup struct {
	ID          int64     `json:"id"`
	LeagueID    int64     `json:"league_id"`
	UserID      int64     `json:"user_id"`
	MatchdayID  int64     `json:"matchday_id"`
	TotalPoints int       `json:"total_points"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LineupPlayer represents a player within a lineup with their assigned position.
type LineupPlayer struct {
	ID        int64          `json:"id"`
	LineupID  int64          `json:"lineup_id"`
	PlayerID  int64          `json:"player_id"`
	Position  PlayerPosition `json:"position"`
	IsStarter bool           `json:"is_starter"`
	Points    int            `json:"points"`
}

// LineupWithPlayers includes the full lineup with player details.
type LineupWithPlayers struct {
	Lineup
	Players []LineupPlayerWithDetails `json:"players"`
}

// LineupPlayerWithDetails combines lineup player with player info.
type LineupPlayerWithDetails struct {
	LineupPlayer
	Player Player `json:"player"`
}

// CreateLineupRequest is the payload for creating/updating a lineup.
type CreateLineupRequest struct {
	MatchdayID int64               `json:"matchday_id" binding:"required"`
	Players    []LineupPlayerInput `json:"players"     binding:"required,min=1"`
}

// LineupPlayerInput is the input for a single player in a lineup.
type LineupPlayerInput struct {
	PlayerID  int64          `json:"player_id" binding:"required"`
	Position  PlayerPosition `json:"position"  binding:"required"`
	IsStarter bool           `json:"is_starter"`
}

// Standings represents the ranking of users in a league.
type Standings struct {
	LeagueID   int64          `json:"league_id"`
	MatchdayID *int64         `json:"matchday_id,omitempty"`
	Rankings   []UserStanding `json:"rankings"`
}

// UserStanding represents a user's position in the standings.
type UserStanding struct {
	Rank        int    `json:"rank"`
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	TotalPoints int    `json:"total_points"`
}
