package market

import (
	"testing"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// md is a matchday with only the date fields the window rule reads.
func md(start, end time.Time) models.Matchday {
	return models.Matchday{StartDate: start, EndDate: end}
}

func TestComputeWindow(t *testing.T) {
	loc, err := Location()
	if err != nil {
		t.Fatalf("load Europe/Madrid: %v", err)
	}

	// Helper to build a local instant in Madrid.
	at := func(y int, m time.Month, d, h, min int) time.Time {
		return time.Date(y, m, d, h, min, 0, 0, loc)
	}

	tests := []struct {
		name       string
		now        time.Time
		matchdays  []models.Matchday
		wantOpen   bool
		wantReason ClosedReason
		wantNext   time.Time
	}{
		{
			name:       "inside window, no matchdays -> open, closes at midnight",
			now:        at(2026, 2, 10, 20, 0),
			matchdays:  nil,
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 2, 11, 0, 0),
		},
		{
			name:       "outside window, no matchdays -> closed, opens at 19:00",
			now:        at(2026, 2, 10, 12, 0),
			matchdays:  nil,
			wantOpen:   false,
			wantReason: ReasonOutsideWindow,
			wantNext:   at(2026, 2, 10, 19, 0),
		},
		{
			name:       "empty matchdays slice behaves like none",
			now:        at(2026, 2, 10, 21, 0),
			matchdays:  []models.Matchday{},
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 2, 11, 0, 0),
		},
		{
			name:       "inside window but matchday in play -> closed until matchday ends",
			now:        at(2026, 2, 10, 20, 0),
			matchdays:  []models.Matchday{md(at(2026, 2, 10, 19, 0), at(2026, 2, 10, 22, 0))},
			wantOpen:   false,
			wantReason: ReasonActiveMatchday,
			wantNext:   at(2026, 2, 10, 22, 0),
		},
		{
			name:       "open now, matchday starts mid-window -> closes when it starts",
			now:        at(2026, 2, 10, 19, 30),
			matchdays:  []models.Matchday{md(at(2026, 2, 10, 21, 0), at(2026, 2, 10, 23, 0))},
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 2, 10, 21, 0),
		},
		{
			name:       "closed by matchday ending outside window -> opens next 19:00",
			now:        at(2026, 2, 10, 20, 0),
			matchdays:  []models.Matchday{md(at(2026, 2, 10, 18, 0), at(2026, 2, 11, 2, 0))},
			wantOpen:   false,
			wantReason: ReasonActiveMatchday,
			wantNext:   at(2026, 2, 11, 19, 0),
		},
		{
			name:       "both conditions closed -> reason is active_matchday (priority)",
			now:        at(2026, 2, 10, 12, 0),
			matchdays:  []models.Matchday{md(at(2026, 2, 10, 10, 0), at(2026, 2, 10, 14, 0))},
			wantOpen:   false,
			wantReason: ReasonActiveMatchday,
			wantNext:   at(2026, 2, 10, 19, 0),
		},
		{
			name:       "edge 19:00 exactly -> open",
			now:        at(2026, 2, 10, 19, 0),
			matchdays:  nil,
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 2, 11, 0, 0),
		},
		{
			name:       "edge 23:59 -> still open, closes at midnight",
			now:        at(2026, 2, 10, 23, 59),
			matchdays:  nil,
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 2, 11, 0, 0),
		},
		{
			name:       "edge midnight 00:00 -> closed, opens at 19:00 same day",
			now:        at(2026, 2, 10, 0, 0),
			matchdays:  nil,
			wantOpen:   false,
			wantReason: ReasonOutsideWindow,
			wantNext:   at(2026, 2, 10, 19, 0),
		},
		{
			// 17:30 UTC in summer = 19:30 CEST (UTC+2) -> open. Proves the named
			// zone is used, not a fixed offset.
			name:       "DST summer: 17:30 UTC is 19:30 CEST -> open",
			now:        time.Date(2026, 7, 1, 17, 30, 0, 0, time.UTC),
			matchdays:  nil,
			wantOpen:   true,
			wantReason: ReasonOpen,
			wantNext:   at(2026, 7, 2, 0, 0),
		},
		{
			// Same 17:30 UTC in winter = 18:30 CET (UTC+1) -> closed. Same UTC
			// instant, different state across the DST boundary.
			name:       "DST winter: 17:30 UTC is 18:30 CET -> closed",
			now:        time.Date(2026, 1, 1, 17, 30, 0, 0, time.UTC),
			matchdays:  nil,
			wantOpen:   false,
			wantReason: ReasonOutsideWindow,
			wantNext:   at(2026, 1, 1, 19, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeWindow(tt.now, loc, tt.matchdays)
			if got.IsOpen != tt.wantOpen {
				t.Errorf("IsOpen = %v, want %v", got.IsOpen, tt.wantOpen)
			}
			if got.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", got.Reason, tt.wantReason)
			}
			if !got.NextChangeAt.Equal(tt.wantNext) {
				t.Errorf("NextChangeAt = %s, want %s",
					got.NextChangeAt.In(loc).Format(time.RFC3339),
					tt.wantNext.In(loc).Format(time.RFC3339))
			}
		})
	}
}
