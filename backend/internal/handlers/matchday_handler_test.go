package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// stubMatchdayRepo satisfies repository.MatchdayRepository. Only the methods the
// standings handlers touch carry behaviour; the rest are no-ops. It exists so
// the standings authorization (A2) can be tested without a real database.
type stubMatchdayRepo struct {
	standings *models.Standings
}

func (s *stubMatchdayRepo) GetStandings(_ context.Context, _ int64, _ *int64) (*models.Standings, error) {
	return s.standings, nil
}
func (s *stubMatchdayRepo) GetByNumber(_ context.Context, _ int) (*models.Matchday, error) {
	return &models.Matchday{ID: 1, Number: 1}, nil
}
func (s *stubMatchdayRepo) Create(_ context.Context, _ *models.Matchday) error { return nil }
func (s *stubMatchdayRepo) GetByID(_ context.Context, _ int64) (*models.Matchday, error) {
	return nil, nil
}
func (s *stubMatchdayRepo) GetAll(_ context.Context) ([]models.Matchday, error)    { return nil, nil }
func (s *stubMatchdayRepo) GetCurrent(_ context.Context) (*models.Matchday, error) { return nil, nil }
func (s *stubMatchdayRepo) Update(_ context.Context, _ *models.Matchday) error     { return nil }
func (s *stubMatchdayRepo) CreateLineup(_ context.Context, _ *models.Lineup) error { return nil }
func (s *stubMatchdayRepo) GetLineup(_ context.Context, _, _, _ int64) (*models.LineupWithPlayers, error) {
	return nil, nil
}
func (s *stubMatchdayRepo) GetDefaultLineup(_ context.Context, _, _ int64) (*models.LineupWithPlayers, error) {
	return nil, nil
}
func (s *stubMatchdayRepo) ReplaceLineupPlayers(_ context.Context, _ int64, _ []models.LineupPlayer) error {
	return nil
}
func (s *stubMatchdayRepo) UpsertLineupPlayer(_ context.Context, _ *models.LineupPlayer) error {
	return nil
}
func (s *stubMatchdayRepo) RemoveLineupPlayer(_ context.Context, _, _ int64) error     { return nil }
func (s *stubMatchdayRepo) UpdateLineupPoints(_ context.Context, _ int64, _ int) error { return nil }

// newMatchdayRouter wires the league-scoped standings endpoints with an
// authenticated caller. It reuses mockLeagueRepoForCreate (from
// league_handler_test.go) as the membership source.
func newMatchdayRouter(t *testing.T, userID int64) (*gin.Engine, *mockLeagueRepoForCreate) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	leagueRepo := &mockLeagueRepoForCreate{}
	h := NewMatchdayHandler(&stubMatchdayRepo{standings: &models.Standings{}}, leagueRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	r.GET("/api/v1/leagues/:id/standings", h.GetStandings)
	r.GET("/api/v1/leagues/:id/matchdays/:number/standings", h.GetMatchdayStandings)
	return r, leagueRepo
}

// A2: standings of a league are only readable by its members.
func TestMatchdayHandler_GetStandings_NonMember_Forbidden(t *testing.T) {
	router, repo := newMatchdayRouter(t, 7)
	repo.getMemberResult = nil // caller is not a member of league 9

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/9/standings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403 para no-miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestMatchdayHandler_GetStandings_Member_OK(t *testing.T) {
	router, repo := newMatchdayRouter(t, 7)
	repo.getMemberResult = &models.LeagueMember{LeagueID: 9, UserID: 7}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/9/standings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200 para miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestMatchdayHandler_GetMatchdayStandings_NonMember_Forbidden(t *testing.T) {
	router, repo := newMatchdayRouter(t, 7)
	repo.getMemberResult = nil

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/9/matchdays/1/standings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403 para no-miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}
