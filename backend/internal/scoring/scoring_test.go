package scoring

import "testing"

func TestComputePoints(t *testing.T) {
	tests := []struct {
		name string
		pos  string
		s    PlayerMatchStats
		want int
	}{
		// ── Minutes tiers ────────────────────────────────────────────────
		{
			name: "no minutes played",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 0},
			want: 0,
		},
		{
			name: "1-59 minutes",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 45},
			want: 1,
		},
		{
			name: "60+ minutes",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90},
			want: 2,
		},
		{
			name: "exactly 60 minutes",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 60},
			want: 2,
		},

		// ── Goals per position ───────────────────────────────────────────
		{
			name: "goal GK",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, Goals: 1},
			want: 2 + 6,
		},
		{
			name: "goal DEF",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, Goals: 1},
			want: 2 + 6,
		},
		{
			name: "goal MID",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90, Goals: 1},
			want: 2 + 5,
		},
		{
			name: "goal FWD",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 90, Goals: 1},
			want: 2 + 4,
		},

		// ── Assist ───────────────────────────────────────────────────────
		{
			name: "assist",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90, Assists: 1},
			want: 2 + 3,
		},
		{
			name: "two assists",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 90, Assists: 2},
			want: 2 + 6,
		},

		// ── Clean sheet per position ─────────────────────────────────────
		{
			name: "clean sheet GK 90min",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, CleanSheet: true},
			want: 2 + 4,
		},
		{
			name: "clean sheet DEF 90min",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, CleanSheet: true},
			want: 2 + 4,
		},
		{
			name: "clean sheet MID 90min",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90, CleanSheet: true},
			want: 2 + 1,
		},
		{
			name: "clean sheet FWD 90min",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 90, CleanSheet: true},
			want: 2,
		},
		{
			name: "clean sheet does not count under 60 min",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 59, CleanSheet: true},
			want: 1,
		},

		// ── Goals conceded (GK/DEF only) ─────────────────────────────────
		{
			name: "GK 2 goals conceded",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 2},
			want: 2 - 1,
		},
		{
			name: "DEF 4 goals conceded",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 4},
			want: 0, // 2 (min) - 2 (4 conceded / 2)
		},
		{
			name: "MID goals conceded no penalty",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 4},
			want: 2,
		},
		{
			name: "FWD goals conceded no penalty",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 4},
			want: 2,
		},
		{
			name: "1 goal conceded no penalty (not a multiple of 2)",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 1},
			want: 2,
		},

		// ── Cards ────────────────────────────────────────────────────────
		{
			name: "yellow card",
			pos:  "MID",
			s:    PlayerMatchStats{Minutes: 90, Yellow: 1},
			want: 2 - 1,
		},
		{
			name: "red card",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, Red: 1},
			want: 2 - 3,
		},

		// ── Own goal ─────────────────────────────────────────────────────
		{
			name: "own goal",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, OwnGoals: 1},
			want: 0, // 2 (min) - 2 (own goal)
		},

		// ── Penalty missed ───────────────────────────────────────────────
		{
			name: "pen missed",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 90, PensMissed: 1},
			want: 0, // 2 (min) - 2 (pen missed)
		},

		// ── GK: penalty saved ────────────────────────────────────────────
		{
			name: "GK pen saved",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, PensSaved: 1},
			want: 2 + 5,
		},
		{
			name: "GK two pens saved",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, PensSaved: 2},
			want: 2 + 10,
		},

		// ── GK: every 3 saves ────────────────────────────────────────────
		{
			name: "GK 3 saves",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, Saves: 3},
			want: 2 + 1,
		},
		{
			name: "GK 6 saves",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, Saves: 6},
			want: 2 + 2,
		},
		{
			name: "GK 2 saves no bonus",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, Saves: 2},
			want: 2,
		},
		{
			name: "non-GK saves give no bonus",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, Saves: 6},
			want: 2,
		},

		// ── Combined real-ish cases ───────────────────────────────────────
		{
			// DEF 90min + 1 goal + clean sheet = 2+6+4 = 12
			name: "DEF 90min goal clean sheet",
			pos:  "DEF",
			s:    PlayerMatchStats{Minutes: 90, Goals: 1, CleanSheet: true},
			want: 12,
		},
		{
			// GK 90min + clean sheet + 6 saves + 0 conceded = 2+4+2 = 8
			name: "GK 90min clean sheet 6 saves",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, CleanSheet: true, Saves: 6, GoalsConceded: 0},
			want: 8,
		},
		{
			// FWD 70min + 1 goal + 1 yellow = 2+4-1 = 5
			name: "FWD 70min goal yellow",
			pos:  "FWD",
			s:    PlayerMatchStats{Minutes: 70, Goals: 1, Yellow: 1},
			want: 5,
		},
		{
			// GK 90min + 2 conceded + 5 saves = 2-1+1 = 2
			name: "GK 90min 2 conceded 5 saves",
			pos:  "GK",
			s:    PlayerMatchStats{Minutes: 90, GoalsConceded: 2, Saves: 5},
			want: 2 - 1 + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputePoints(tt.pos, tt.s)
			if got != tt.want {
				t.Errorf("ComputePoints(%q, %+v) = %d, want %d", tt.pos, tt.s, got, tt.want)
			}
		})
	}
}
