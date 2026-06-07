package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

func teamRouter(userID int64) *gin.Engine {
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

func doTeamReq(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── GetUserTeam tests ────────────────────────────────────────────────────────

func TestTeam_GetUserTeam_NoAuth(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := teamRouter(0)
	r.GET("/leagues/:id/team", h.GetUserTeam)
	w := doTeamReq(r, "GET", "/leagues/1/team")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestTeam_GetUserTeam_InvalidLeagueID(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)
	w := doTeamReq(r, "GET", "/leagues/abc/team")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTeam_GetUserTeam_NotMember(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return nil, nil
		},
	}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)
	w := doTeamReq(r, "GET", "/leagues/1/team")
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestTeam_GetUserTeam_TeamNotFound(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return &models.LeagueMember{}, nil
		},
	}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)
	w := doTeamReq(r, "GET", "/leagues/1/team")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ── GetUserTeamInLeague tests ────────────────────────────────────────────────

func TestTeam_GetUserTeamInLeague_NoAuth(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := teamRouter(0)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)
	w := doTeamReq(r, "GET", "/leagues/1/users/2/team")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestTeam_GetUserTeamInLeague_CallerNotMember(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, userID int64) (*models.LeagueMember, error) {
			if userID == 1 {
				return nil, nil
			}
			return &models.LeagueMember{}, nil
		},
	}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)
	w := doTeamReq(r, "GET", "/leagues/1/users/2/team")
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestTeam_GetUserTeamInLeague_TargetNotMember(t *testing.T) {
	callCount := 0
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			callCount++
			if callCount == 1 {
				return &models.LeagueMember{}, nil
			}
			return nil, nil
		},
	}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)
	w := doTeamReq(r, "GET", "/leagues/1/users/999/team")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestTeam_GetUserTeamInLeague_InvalidUserID(t *testing.T) {
	h := NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := teamRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)
	w := doTeamReq(r, "GET", "/leagues/1/users/abc/team")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
