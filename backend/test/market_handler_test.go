package test

import (
	"net/http"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
)

// ── Auth guard tests ─────────────────────────────────────────────────────────

func TestGetAvailablePlayers_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)

	w := doReq(r, "GET", "/leagues/1/market/players", nil)
	assertStatus(t, w, http.StatusUnauthorized)
	assertBodyContains(t, w, "unauthorized")
}

func TestGetActiveListings_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)

	w := doReq(r, "GET", "/leagues/1/market/listings", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestPlaceBid_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)

	w := doReq(r, "POST", "/leagues/1/market/bids", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetUserBids_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/market/bids", h.GetUserBids)

	w := doReq(r, "GET", "/leagues/1/market/bids", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestCancelBid_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)

	w := doReq(r, "DELETE", "/leagues/1/market/bids/1", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMarketStatus_NoAuth(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)

	w := doReq(r, "GET", "/leagues/1/market/status", nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

// ── Invalid league ID tests ──────────────────────────────────────────────────

func TestGetAvailablePlayers_InvalidLeagueID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)

	w := doReq(r, "GET", "/leagues/abc/market/players", nil)
	assertStatus(t, w, http.StatusBadRequest)
	assertBodyContains(t, w, "Invalid league ID")
}

func TestGetActiveListings_InvalidLeagueID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)

	w := doReq(r, "GET", "/leagues/xyz/market/listings", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestPlaceBid_InvalidLeagueID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)

	w := doReq(r, "POST", "/leagues/abc/market/bids", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCancelBid_InvalidLeagueID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)

	w := doReq(r, "DELETE", "/leagues/abc/market/bids/1", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestCancelBid_InvalidBidID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)

	w := doReq(r, "DELETE", "/leagues/notanumber/market/bids/abc", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

func TestGetMarketStatus_InvalidLeagueID(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(1)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)

	w := doReq(r, "GET", "/leagues/abc/market/status", nil)
	assertStatus(t, w, http.StatusBadRequest)
}

// ── PlaceBid: auth guard fires before body parse ─────────────────────────────

func TestPlaceBid_BadBody_ButNoAuthFirst(t *testing.T) {
	h := handlers.NewMarketHandler(nil, nil, nil, nil)
	r := newRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)

	w := doReq(r, "POST", "/leagues/1/market/bids", map[string]string{"bad": "body"})
	assertStatus(t, w, http.StatusUnauthorized)
}
