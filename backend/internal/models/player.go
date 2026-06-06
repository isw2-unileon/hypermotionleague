package models

import "time"

// PlayerPosition defines the position on the field.
type PlayerPosition string

const (
	PositionGK  PlayerPosition = "GK"
	PositionDEF PlayerPosition = "DEF"
	PositionMID PlayerPosition = "MID"
	PositionFWD PlayerPosition = "FWD"
)

// Player represents a football player in the global pool.
type Player struct {
	ID          int64          `json:"id"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Position    PlayerPosition `json:"position"`
	TeamName    string         `json:"team_name"`
	MarketValue int            `json:"market_value"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// API-Football sync fields, written by the sync-players command. The
	// existing read queries do not select these, so they stay zero/nil on
	// normal reads (and are omitted from JSON). JerseyNumber and Age are
	// pointers so an unknown value (API null / 0) is stored as SQL NULL.
	ExternalID   int64  `json:"external_id,omitempty"` // API-Football player.id
	TeamID       *int64 `json:"team_id,omitempty"`     // FK to teams.id (real club)
	JerseyNumber *int   `json:"jersey_number,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
	Age          *int   `json:"age,omitempty"`
}

// FullName returns the player's full name.
func (p Player) FullName() string {
	return p.FirstName + " " + p.LastName
}

// PlayerPoints represents a player's stats and points for a matchday.
type PlayerPoints struct {
	ID            int64 `json:"id"`
	PlayerID      int64 `json:"player_id"`
	MatchdayID    int64 `json:"matchday_id"`
	Points        int   `json:"points"`
	Goals         int   `json:"goals"`
	Assists       int   `json:"assists"`
	MinutesPlayed int   `json:"minutes_played"`
	YellowCards   int   `json:"yellow_cards"`
	RedCards      int   `json:"red_cards"`
	CleanSheet    bool  `json:"clean_sheet"`
	// Stat columns added in migration 004_extend_player_points_stats, written by
	// the ingest pipeline so the scoring inputs persist alongside the points.
	GoalsConceded int       `json:"goals_conceded"`
	PensMissed    int       `json:"pens_missed"`
	PensSaved     int       `json:"pens_saved"`
	Saves         int       `json:"saves"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PlayerWithPoints combines player data with their points for a matchday.
type PlayerWithPoints struct {
	Player
	Points *PlayerPoints `json:"points,omitempty"`
}
