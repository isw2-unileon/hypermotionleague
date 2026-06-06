package handlers_test

import (
	"net/http"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
	h "github.com/isw2-unileon/proyect-scaffolding/backend/test/helpers"
)

// ── Auth guard tests ─────────────────────────────────────────────────────────

func TestGetAvailablePlayers_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.GET("/leagues/:id/market/players", mh.GetAvailablePlayers)

	w := h.DoReq(r, "GET", "/leagues/1/market/players", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
	h.AssertBodyContains(t, w, "unauthorized")
}

func TestGetActiveListings_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.GET("/leagues/:id/market/listings", mh.GetActiveListings)

	w := h.DoReq(r, "GET", "/leagues/1/market/listings", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestPlaceBid_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.POST("/leagues/:id/market/bids", mh.PlaceBid)

	w := h.DoReq(r, "POST", "/leagues/1/market/bids", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestGetUserBids_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.GET("/leagues/:id/market/bids", mh.GetUserBids)

	w := h.DoReq(r, "GET", "/leagues/1/market/bids", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestCancelBid_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", mh.CancelBid)

	w := h.DoReq(r, "DELETE", "/leagues/1/market/bids/1", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestGetMarketStatus_NoAuth(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.GET("/leagues/:id/market/status", mh.GetMarketStatus)

	w := h.DoReq(r, "GET", "/leagues/1/market/status", nil)
	h.AssertStatus(t, w, http.StatusUnauthorized)
}

// ── Invalid league ID tests ──────────────────────────────────────────────────

func TestGetAvailablePlayers_InvalidLeagueID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.GET("/leagues/:id/market/players", mh.GetAvailablePlayers)

	w := h.DoReq(r, "GET", "/leagues/abc/market/players", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
	h.AssertBodyContains(t, w, "Invalid league ID")
}

func TestGetActiveListings_InvalidLeagueID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.GET("/leagues/:id/market/listings", mh.GetActiveListings)

	w := h.DoReq(r, "GET", "/leagues/xyz/market/listings", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestPlaceBid_InvalidLeagueID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.POST("/leagues/:id/market/bids", mh.PlaceBid)

	w := h.DoReq(r, "POST", "/leagues/abc/market/bids", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestCancelBid_InvalidLeagueID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.DELETE("/leagues/:id/market/bids/:bid_id", mh.CancelBid)

	w := h.DoReq(r, "DELETE", "/leagues/abc/market/bids/1", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestCancelBid_InvalidBidID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.DELETE("/leagues/:id/market/bids/:bid_id", mh.CancelBid)

	w := h.DoReq(r, "DELETE", "/leagues/notanumber/market/bids/abc", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestGetMarketStatus_InvalidLeagueID(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(1)
	r.GET("/leagues/:id/market/status", mh.GetMarketStatus)

	w := h.DoReq(r, "GET", "/leagues/abc/market/status", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

// ── PlaceBid: auth guard fires before body parse ─────────────────────────────

func TestPlaceBid_BadBody_ButNoAuthFirst(t *testing.T) {
	mh := handlers.NewMarketHandler(nil, nil, nil, nil, nil)
	r := h.NewRouter(0)
	r.POST("/leagues/:id/market/bids", mh.PlaceBid)

	w := h.DoReq(r, "POST", "/leagues/1/market/bids", map[string]string{"bad": "body"})
	h.AssertStatus(t, w, http.StatusUnauthorized)
}
