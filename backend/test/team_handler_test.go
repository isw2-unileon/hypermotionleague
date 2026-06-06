package test

import (
	"context"
	"net/http"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// ── GetUserTeam tests ────────────────────────────────────────────────────────

func TestGetUserTeam_NoAuth(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/team", h.GetUserTeam)

	w := doReq(r, "GET", "/leagues/1/team", nil)
	assertStatusCode(t, w, http.StatusUnauthorized)
}

func TestGetUserTeam_InvalidLeagueID(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)

	w := doReq(r, "GET", "/leagues/abc/team", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestGetUserTeam_NotMember(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return nil, nil
		},
	}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)

	w := doReq(r, "GET", "/leagues/1/team", nil)
	assertStatusCode(t, w, http.StatusForbidden)
}

func TestGetUserTeam_TeamNotFound(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return &models.LeagueMember{}, nil
		},
	}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/team", h.GetUserTeam)

	w := doReq(r, "GET", "/leagues/1/team", nil)
	assertStatusCode(t, w, http.StatusNotFound)
}

// ── GetUserTeamInLeague tests ────────────────────────────────────────────────

func TestGetUserTeamInLeague_NoAuth(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)

	w := doReq(r, "GET", "/leagues/1/users/2/team", nil)
	assertStatusCode(t, w, http.StatusUnauthorized)
}

func TestGetUserTeamInLeague_CallerNotMember(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, userID int64) (*models.LeagueMember, error) {
			if userID == 1 {
				return nil, nil
			}
			return &models.LeagueMember{}, nil
		},
	}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)

	w := doReq(r, "GET", "/leagues/1/users/2/team", nil)
	assertStatusCode(t, w, http.StatusForbidden)
}

func TestGetUserTeamInLeague_TargetNotMember(t *testing.T) {
	callCount := 0
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			callCount++
			if callCount == 1 {
				return &models.LeagueMember{}, nil
			}
			return nil, nil
		},
	}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)

	w := doReq(r, "GET", "/leagues/1/users/999/team", nil)
	assertStatusCode(t, w, http.StatusNotFound)
}

func TestGetUserTeamInLeague_InvalidUserID(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/users/:userId/team", h.GetUserTeamInLeague)

	w := doReq(r, "GET", "/leagues/1/users/abc/team", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

// ── BuyPlayer tests ──────────────────────────────────────────────────────────

func TestBuyPlayer_NoAuth(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(0)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/1/users/2/team/10/buy", nil)
	assertStatusCode(t, w, http.StatusUnauthorized)
}

func TestBuyPlayer_InvalidLeagueID(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/abc/users/2/team/10/buy", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestBuyPlayer_InvalidSellerID(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/1/users/abc/team/10/buy", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestBuyPlayer_InvalidPlayerID(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/1/users/2/team/abc/buy", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestBuyPlayer_CannotBuyOwnPlayer(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{}, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/1/users/1/team/10/buy", nil)
	assertStatusCode(t, w, http.StatusBadRequest)
}

func TestBuyPlayer_NotMember(t *testing.T) {
	h := handlers.NewTeamHandler(&mockTeamRepo{}, &mockLeagueRepo{
		fnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return nil, nil
		},
	}, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/users/:userId/team/:playerId/buy", h.BuyPlayer)

	w := doReq(r, "POST", "/leagues/1/users/2/team/10/buy", nil)
	assertStatusCode(t, w, http.StatusForbidden)
}
