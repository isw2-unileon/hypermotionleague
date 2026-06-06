package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	h "github.com/isw2-unileon/proyect-scaffolding/backend/test/helpers"
)

// ── GetUserTeam tests ────────────────────────────────────────────────────────

func TestGetUserTeam_NoAuth(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{})
	r := h.NewRouter(0)
	r.GET("/leagues/:id/team", th.GetUserTeam)

	w := h.DoReq(r, "GET", "/leagues/1/team", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestGetUserTeam_InvalidLeagueID(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/team", th.GetUserTeam)

	w := h.DoReq(r, "GET", "/leagues/abc/team", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestGetUserTeam_NotMember(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{
		FnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return nil, nil
		},
	})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/team", th.GetUserTeam)

	w := h.DoReq(r, "GET", "/leagues/1/team", nil)
	h.AssertStatus(t, w, http.StatusForbidden)
}

func TestGetUserTeam_TeamNotFound(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{
		FnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			return &models.LeagueMember{}, nil
		},
	})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/team", th.GetUserTeam)

	w := h.DoReq(r, "GET", "/leagues/1/team", nil)
	h.AssertStatus(t, w, http.StatusNotFound)
}

// ── GetUserTeamInLeague tests ────────────────────────────────────────────────

func TestGetUserTeamInLeague_NoAuth(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{})
	r := h.NewRouter(0)
	r.GET("/leagues/:id/users/:userId/team", th.GetUserTeamInLeague)

	w := h.DoReq(r, "GET", "/leagues/1/users/2/team", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestGetUserTeamInLeague_CallerNotMember(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{
		FnGetMember: func(_ context.Context, _, userID int64) (*models.LeagueMember, error) {
			if userID == 1 {
				return nil, nil
			}
			return &models.LeagueMember{}, nil
		},
	})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/users/:userId/team", th.GetUserTeamInLeague)

	w := h.DoReq(r, "GET", "/leagues/1/users/2/team", nil)
	h.AssertStatus(t, w, http.StatusForbidden)
}

func TestGetUserTeamInLeague_TargetNotMember(t *testing.T) {
	callCount := 0
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{
		FnGetMember: func(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
			callCount++
			if callCount == 1 {
				return &models.LeagueMember{}, nil
			}
			return nil, nil
		},
	})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/users/:userId/team", th.GetUserTeamInLeague)

	w := h.DoReq(r, "GET", "/leagues/1/users/999/team", nil)
	h.AssertStatus(t, w, http.StatusNotFound)
}

func TestGetUserTeamInLeague_InvalidUserID(t *testing.T) {
	th := handlers.NewTeamHandler(&h.MockTeamRepo{}, &h.MockLeagueRepo{})
	r := h.NewRouter(1)
	r.GET("/leagues/:id/users/:userId/team", th.GetUserTeamInLeague)

	w := h.DoReq(r, "GET", "/leagues/1/users/abc/team", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}
