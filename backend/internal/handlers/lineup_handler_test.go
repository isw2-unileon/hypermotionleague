package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// ── mock repositories ──────────────────────────────────────────────────────

type mockLeagueRepo struct {
	fnGetMember func(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error)
}

func (m *mockLeagueRepo) Create(_ context.Context, _ *models.League) error { return nil }
func (m *mockLeagueRepo) GetByID(_ context.Context, _ int64) (*models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetByInviteCode(_ context.Context, _ string) (*models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetByUserID(_ context.Context, _ int64) ([]models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) Update(_ context.Context, _ *models.League) error  { return nil }
func (m *mockLeagueRepo) Delete(_ context.Context, _ int64) error            { return nil }
func (m *mockLeagueRepo) AddMember(_ context.Context, _ *models.LeagueMember) error {
	return nil
}
func (m *mockLeagueRepo) GetMembers(_ context.Context, _ int64) ([]models.LeagueMember, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetMembersWithUsers(_ context.Context, _ int64) ([]models.LeagueMemberWithUser, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetMember(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error) {
	if m.fnGetMember != nil {
		return m.fnGetMember(ctx, leagueID, userID)
	}
	return &models.LeagueMember{}, nil
}
func (m *mockLeagueRepo) UpdateMemberBudget(_ context.Context, _, _ int64, _ int) error {
	return nil
}
func (m *mockLeagueRepo) RemoveMember(_ context.Context, _, _ int64) error { return nil }
func (m *mockLeagueRepo) CountMembers(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

type mockTeamRepo struct {
	fnHasPlayer func(ctx context.Context, leagueID, userID, playerID int64) (bool, error)
}

func (m *mockTeamRepo) AddPlayer(_ context.Context, _ *models.TeamPlayer) error { return nil }
func (m *mockTeamRepo) RemovePlayer(_ context.Context, _, _, _ int64) error      { return nil }
func (m *mockTeamRepo) GetUserTeam(_ context.Context, _, _ int64) (*models.UserTeam, error) {
	return nil, nil
}
func (m *mockTeamRepo) GetPlayerOwner(_ context.Context, _, _ int64) (*models.TeamPlayer, error) {
	return nil, nil
}
func (m *mockTeamRepo) HasPlayer(ctx context.Context, leagueID, userID, playerID int64) (bool, error) {
	if m.fnHasPlayer != nil {
		return m.fnHasPlayer(ctx, leagueID, userID, playerID)
	}
	return true, nil
}
func (m *mockTeamRepo) TransferPlayer(_ context.Context, _, _, _, _ int64, _ int) error {
	return nil
}

type mockMatchdayRepo struct {
	fnGetByLeague          func(ctx context.Context, leagueID int64) ([]models.Matchday, error)
	fnGetLineup            func(ctx context.Context, leagueID, userID, matchdayID int64) (*models.LineupWithPlayers, error)
	fnCreateLineup         func(ctx context.Context, lineup *models.Lineup) error
	fnReplaceLineupPlayers func(ctx context.Context, lineupID int64, players []models.LineupPlayer) error
	fnRemoveLineupPlayer   func(ctx context.Context, lineupID, playerID int64) error
}

func (m *mockMatchdayRepo) Create(_ context.Context, _ *models.Matchday) error { return nil }
func (m *mockMatchdayRepo) GetByID(_ context.Context, _ int64) (*models.Matchday, error) {
	return nil, nil
}
func (m *mockMatchdayRepo) GetByLeague(ctx context.Context, leagueID int64) ([]models.Matchday, error) {
	if m.fnGetByLeague != nil {
		return m.fnGetByLeague(ctx, leagueID)
	}
	return nil, nil
}
func (m *mockMatchdayRepo) GetCurrent(_ context.Context, _ int64) (*models.Matchday, error) {
	return nil, nil
}
func (m *mockMatchdayRepo) Update(_ context.Context, _ *models.Matchday) error { return nil }
func (m *mockMatchdayRepo) CreateLineup(ctx context.Context, lineup *models.Lineup) error {
	if m.fnCreateLineup != nil {
		return m.fnCreateLineup(ctx, lineup)
	}
	return nil
}
func (m *mockMatchdayRepo) GetLineup(ctx context.Context, leagueID, userID, matchdayID int64) (*models.LineupWithPlayers, error) {
	if m.fnGetLineup != nil {
		return m.fnGetLineup(ctx, leagueID, userID, matchdayID)
	}
	return &models.LineupWithPlayers{Lineup: models.Lineup{ID: 1}}, nil
}
func (m *mockMatchdayRepo) ReplaceLineupPlayers(ctx context.Context, lineupID int64, players []models.LineupPlayer) error {
	if m.fnReplaceLineupPlayers != nil {
		return m.fnReplaceLineupPlayers(ctx, lineupID, players)
	}
	return nil
}
func (m *mockMatchdayRepo) UpsertLineupPlayer(_ context.Context, _ *models.LineupPlayer) error {
	return nil
}
func (m *mockMatchdayRepo) RemoveLineupPlayer(ctx context.Context, lineupID, playerID int64) error {
	if m.fnRemoveLineupPlayer != nil {
		return m.fnRemoveLineupPlayer(ctx, lineupID, playerID)
	}
	return nil
}
func (m *mockMatchdayRepo) UpdateLineupPoints(_ context.Context, _ int64, _ int) error {
	return nil
}
func (m *mockMatchdayRepo) GetStandings(_ context.Context, _ int64, _ *int64) (*models.Standings, error) {
	return nil, nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func newTestRouter(h *LineupHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", int64(1))
		c.Next()
	})
	r.GET("/leagues/:id/matchdays/:number/lineup", h.GetLineup)
	r.PUT("/leagues/:id/matchdays/:number/lineup", h.SaveLineup)
	r.DELETE("/leagues/:id/matchdays/:number/lineup/players/:player_id", h.RemoveLineupPlayer)
	return r
}

func futureMatchday() models.Matchday {
	return models.Matchday{ID: 10, LeagueID: 1, Number: 1, StartDate: time.Now().Add(24 * time.Hour)}
}

// valid442 builds an 11-player 4-4-2 request (1 GK, 4 DEF, 4 MID, 2 FWD).
func valid442() models.CreateLineupRequest {
	return models.CreateLineupRequest{
		MatchdayID: 10,
		Players: []models.LineupPlayerInput{
			{PlayerID: 1, Position: models.PositionGK, IsStarter: true},
			{PlayerID: 2, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 3, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 4, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 5, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 6, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 7, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 8, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 9, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 10, Position: models.PositionFWD, IsStarter: true},
			{PlayerID: 11, Position: models.PositionFWD, IsStarter: true},
		},
	}
}

func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── SaveLineup tests ───────────────────────────────────────────────────────

func TestSaveLineup_DeadlinePassed(t *testing.T) {
	md := futureMatchday()
	md.StartDate = time.Now().Add(-1 * time.Hour)

	h := NewLineupHandler(
		&mockMatchdayRepo{fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
			return []models.Matchday{md}, nil
		}},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", valid442())
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSaveLineup_InvalidFormation_MultipleGK(t *testing.T) {
	md := futureMatchday()
	req := models.CreateLineupRequest{
		MatchdayID: 10,
		Players: []models.LineupPlayerInput{
			{PlayerID: 1, Position: models.PositionGK, IsStarter: true},
			{PlayerID: 2, Position: models.PositionGK, IsStarter: true},
			{PlayerID: 3, Position: models.PositionGK, IsStarter: true},
			{PlayerID: 4, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 5, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 6, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 7, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 8, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 9, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 10, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 11, Position: models.PositionMID, IsStarter: true},
		},
	}

	h := NewLineupHandler(
		&mockMatchdayRepo{fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
			return []models.Matchday{md}, nil
		}},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSaveLineup_InvalidFormation_TooFewDEF(t *testing.T) {
	md := futureMatchday()
	// 1 GK + 2 DEF + 5 MID + 3 FWD = 11, but DEF < 3
	req := models.CreateLineupRequest{
		MatchdayID: 10,
		Players: []models.LineupPlayerInput{
			{PlayerID: 1, Position: models.PositionGK, IsStarter: true},
			{PlayerID: 2, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 3, Position: models.PositionDEF, IsStarter: true},
			{PlayerID: 4, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 5, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 6, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 7, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 8, Position: models.PositionMID, IsStarter: true},
			{PlayerID: 9, Position: models.PositionFWD, IsStarter: true},
			{PlayerID: 10, Position: models.PositionFWD, IsStarter: true},
			{PlayerID: 11, Position: models.PositionFWD, IsStarter: true},
		},
	}

	h := NewLineupHandler(
		&mockMatchdayRepo{fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
			return []models.Matchday{md}, nil
		}},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSaveLineup_ValidFormation_442(t *testing.T) {
	md := futureMatchday()

	h := NewLineupHandler(
		&mockMatchdayRepo{fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
			return []models.Matchday{md}, nil
		}},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", valid442())
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSaveLineup_StalePlayersReplaced(t *testing.T) {
	md := futureMatchday()
	var capturedPlayers []models.LineupPlayer

	h := NewLineupHandler(
		&mockMatchdayRepo{
			fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
				return []models.Matchday{md}, nil
			},
			fnReplaceLineupPlayers: func(_ context.Context, _ int64, players []models.LineupPlayer) error {
				capturedPlayers = players
				return nil
			},
		},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	req := valid442()
	doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", req)

	if len(capturedPlayers) != len(req.Players) {
		t.Errorf("ReplaceLineupPlayers called with %d players, want %d", len(capturedPlayers), len(req.Players))
	}
}

func TestSaveLineup_NotMember(t *testing.T) {
	h := NewLineupHandler(
		&mockMatchdayRepo{},
		&mockTeamRepo{},
		&mockLeagueRepo{fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return nil, nil
		}},
	)

	w := doRequest(newTestRouter(h), http.MethodPut, "/leagues/1/matchdays/1/lineup", valid442())
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// ── GetLineup tests ────────────────────────────────────────────────────────

func TestGetLineup_NotFound(t *testing.T) {
	md := futureMatchday()

	h := NewLineupHandler(
		&mockMatchdayRepo{
			fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
				return []models.Matchday{md}, nil
			},
			fnGetLineup: func(_ context.Context, _, _, _ int64) (*models.LineupWithPlayers, error) {
				return nil, nil
			},
		},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodGet, "/leagues/1/matchdays/1/lineup", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetLineup_Success(t *testing.T) {
	md := futureMatchday()
	lineup := &models.LineupWithPlayers{Lineup: models.Lineup{ID: 1, LeagueID: 1, UserID: 1}}

	h := NewLineupHandler(
		&mockMatchdayRepo{
			fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
				return []models.Matchday{md}, nil
			},
			fnGetLineup: func(_ context.Context, _, _, _ int64) (*models.LineupWithPlayers, error) {
				return lineup, nil
			},
		},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodGet, "/leagues/1/matchdays/1/lineup", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ── RemoveLineupPlayer tests ───────────────────────────────────────────────

func TestRemoveLineupPlayer_Success(t *testing.T) {
	md := futureMatchday()
	removed := false

	h := NewLineupHandler(
		&mockMatchdayRepo{
			fnGetByLeague: func(_ context.Context, _ int64) ([]models.Matchday, error) {
				return []models.Matchday{md}, nil
			},
			fnRemoveLineupPlayer: func(_ context.Context, _, _ int64) error {
				removed = true
				return nil
			},
		},
		&mockTeamRepo{},
		&mockLeagueRepo{},
	)

	w := doRequest(newTestRouter(h), http.MethodDelete, "/leagues/1/matchdays/1/lineup/players/5", nil)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if !removed {
		t.Error("RemoveLineupPlayer was not called")
	}
}
