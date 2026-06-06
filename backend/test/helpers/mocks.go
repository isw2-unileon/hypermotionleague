// Package helpers provides shared mocks and utilities for handler tests.
package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// ── Mock repositories ────────────────────────────────────────────────────────

// MockLeagueRepo implements repository.LeagueRepository for testing.
type MockLeagueRepo struct {
	FnGetMember func(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error)
}

func (m *MockLeagueRepo) Create(_ context.Context, _ *models.League) error     { return nil }
func (m *MockLeagueRepo) GetByID(_ context.Context, _ int64) (*models.League, error) {
	return nil, nil
}
func (m *MockLeagueRepo) GetByInviteCode(_ context.Context, _ string) (*models.League, error) {
	return nil, nil
}
func (m *MockLeagueRepo) GetByUserID(_ context.Context, _ int64) ([]models.League, error) {
	return nil, nil
}
func (m *MockLeagueRepo) Update(_ context.Context, _ *models.League) error { return nil }
func (m *MockLeagueRepo) Delete(_ context.Context, _ int64) error          { return nil }
func (m *MockLeagueRepo) AddMember(_ context.Context, _ *models.LeagueMember) error {
	return nil
}
func (m *MockLeagueRepo) GetMembers(_ context.Context, _ int64) ([]models.LeagueMember, error) {
	return nil, nil
}
func (m *MockLeagueRepo) GetMembersWithUsers(_ context.Context, _ int64) ([]models.LeagueMemberWithUser, error) {
	return nil, nil
}
func (m *MockLeagueRepo) GetMember(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error) {
	if m.FnGetMember != nil {
		return m.FnGetMember(ctx, leagueID, userID)
	}
	return &models.LeagueMember{}, nil
}
func (m *MockLeagueRepo) UpdateMemberBudget(_ context.Context, _, _ int64, _ int) error {
	return nil
}
func (m *MockLeagueRepo) RemoveMember(_ context.Context, _, _ int64) error { return nil }
func (m *MockLeagueRepo) CountMembers(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

// MockTeamRepo implements repository.TeamRepository for testing.
type MockTeamRepo struct{}

func (m *MockTeamRepo) AddPlayer(_ context.Context, _ *models.TeamPlayer) error { return nil }
func (m *MockTeamRepo) RemovePlayer(_ context.Context, _, _, _ int64) error     { return nil }
func (m *MockTeamRepo) GetUserTeam(_ context.Context, _, _ int64) (*models.UserTeam, error) {
	return nil, nil
}
func (m *MockTeamRepo) GetPlayerOwner(_ context.Context, _, _ int64) (*models.TeamPlayer, error) {
	return nil, nil
}
func (m *MockTeamRepo) HasPlayer(_ context.Context, _, _, _ int64) (bool, error) {
	return true, nil
}
func (m *MockTeamRepo) TransferPlayer(_ context.Context, _, _, _, _ int64, _ int) error {
	return nil
}
func (m *MockTeamRepo) DraftInitialSquad(_ context.Context, _, _ int64) error {
	return nil
}

// ── Test helpers ─────────────────────────────────────────────────────────────

// NewRouter creates a gin test router with the given userID injected into context.
// Pass 0 for unauthenticated requests.
func NewRouter(userID int64) *gin.Engine {
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

// DoReq performs an HTTP request against the given router and returns the response.
func DoReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

// AssertStatus fails the test if the response status code doesn't match.
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}

// AssertBodyContains fails the test if the response body doesn't contain the substring.
func AssertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	if !bytes.Contains(w.Body.Bytes(), []byte(substr)) {
		t.Errorf("body %q does not contain %q", w.Body.String(), substr)
	}
}
