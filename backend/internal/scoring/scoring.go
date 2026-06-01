package scoring

// PlayerMatchStats holds the raw per-match stats used to compute fantasy points.
type PlayerMatchStats struct {
	Minutes       int
	Goals         int
	Assists       int
	OwnGoals      int
	Yellow        int
	Red           int
	GoalsConceded int
	CleanSheet    bool
	PensMissed    int
	PensSaved     int
	Saves         int
}

// ComputePoints returns the fantasy points for a player given their position and match stats.
// pos must be one of "GK", "DEF", "MID", "FWD".
func ComputePoints(pos string, s PlayerMatchStats) int {
	pts := 0

	// Minutes played
	if s.Minutes >= 60 {
		pts += 2
	} else if s.Minutes >= 1 {
		pts += 1
	}

	// Goals scored
	switch pos {
	case "GK", "DEF":
		pts += s.Goals * 6
	case "MID":
		pts += s.Goals * 5
	case "FWD":
		pts += s.Goals * 4
	}

	// Assists
	pts += s.Assists * 3

	// Clean sheet (only counts when player was on the pitch >= 60 min)
	if s.CleanSheet && s.Minutes >= 60 {
		switch pos {
		case "GK", "DEF":
			pts += 4
		case "MID":
			pts += 1
		// FWD: 0
		}
	}

	// Goals conceded: -1 per every 2 conceded (GK and DEF only)
	if pos == "GK" || pos == "DEF" {
		pts -= s.GoalsConceded / 2
	}

	// Cards
	pts -= s.Yellow     // -1 per yellow
	pts -= s.Red * 3    // -3 per red

	// Own goals and penalty misses: -2 each
	pts -= s.OwnGoals * 2
	pts -= s.PensMissed * 2

	// GK-only bonuses
	if pos == "GK" {
		pts += s.PensSaved * 5
		pts += s.Saves / 3
	}

	return pts
}
