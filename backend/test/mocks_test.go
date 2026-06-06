package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// ── mock repositories ────────────────────────────────────────────────────────

type mockLeagueRepo struct {
	fnGetMember func(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error)
}

func (m *mockLeagueRepo) Create(_ context.Context, _ *models.League) error     { return nil }
func (m *mockLeagueRepo) GetByID(_ context.Context, _ int64) (*models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetByInviteCode(_ context.Context, _ string) (*models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) GetByUserID(_ context.Context, _ int64) ([]models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepo) Update(_ context.Context, _ *models.League) error     { return nil }
func (m *mockLeagueRepo) Delete(_ context.Context, _ int64) error              { return nil }
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

type mockTeamRepo struct{}

func (m *mockTeamRepo) AddPlayer(_ context.Context, _ *models.TeamPlayer) error { return nil }
func (m *mockTeamRepo) RemovePlayer(_ context.Context, _, _, _ int64) error     { return nil }
func (m *mockTeamRepo) GetUserTeam(_ context.Context, _, _ int64) (*models.UserTeam, error) {
	return nil, nil
}
func (m *mockTeamRepo) GetPlayerOwner(_ context.Context, _, _ int64) (*models.TeamPlayer, error) {
	return nil, nil
}
func (m *mockTeamRepo) HasPlayer(_ context.Context, _, _, _ int64) (bool, error) {
	return true, nil
}
func (m *mockTeamRepo) TransferPlayer(_ context.Context, _, _, _, _ int64, _ int) error {
	return nil
}
func (m *mockTeamRepo) DraftInitialSquad(_ context.Context, _, _ int64) error {
	return nil
}

// ── test helpers ─────────────────────────────────────────────────────────────

func newRouter(userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if userID != 0 {
			c.Set("userID", userID)
		}
		c.Next()
	})
	return r
}

func doReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}

func assertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	if !bytes.Contains(w.Body.Bytes(), []byte(substr)) {
		t.Errorf("body %q does not contain %q", w.Body.String(), substr)
	}
}

func assertStatusCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}
