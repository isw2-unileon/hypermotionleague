package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func marketRouter(userID int64) *gin.Engine {
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

func doMarketReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
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

func assertMarketStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}

func assertMarketBody(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	if !bytes.Contains(w.Body.Bytes(), []byte(substr)) {
		t.Errorf("body %q does not contain %q", w.Body.String(), substr)
	}
}

// ── Auth guard tests ─────────────────────────────────────────────────────────

func TestMarket_GetAvailablePlayers_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)
	w := doMarketReq(r, "GET", "/leagues/1/market/players", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
	assertMarketBody(t, w, "unauthorized")
}

func TestMarket_GetActiveListings_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)
	w := doMarketReq(r, "GET", "/leagues/1/market/listings", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_PlaceBid_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/1/market/bids", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_GetUserBids_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.GET("/leagues/:id/market/bids", h.GetUserBids)
	w := doMarketReq(r, "GET", "/leagues/1/market/bids", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_CancelBid_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)
	w := doMarketReq(r, "DELETE", "/leagues/1/market/bids/1", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_GetMarketStatus_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)
	w := doMarketReq(r, "GET", "/leagues/1/market/status", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

// ── Invalid league ID tests ──────────────────────────────────────────────────

func TestMarket_GetAvailablePlayers_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(1)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)
	w := doMarketReq(r, "GET", "/leagues/abc/market/players", nil)
	assertMarketStatus(t, w, http.StatusBadRequest)
	assertMarketBody(t, w, "Invalid league ID")
}

func TestMarket_GetActiveListings_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(1)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)
	w := doMarketReq(r, "GET", "/leagues/xyz/market/listings", nil)
	assertMarketStatus(t, w, http.StatusBadRequest)
}

func TestMarket_PlaceBid_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(1)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/abc/market/bids", nil)
	assertMarketStatus(t, w, http.StatusBadRequest)
}

func TestMarket_CancelBid_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)
	w := doMarketReq(r, "DELETE", "/leagues/abc/market/bids/1", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_CancelBid_InvalidBidID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)
	w := doMarketReq(r, "DELETE", "/leagues/notanumber/market/bids/abc", nil)
	assertMarketStatus(t, w, http.StatusUnauthorized)
}

func TestMarket_GetMarketStatus_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(1)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)
	w := doMarketReq(r, "GET", "/leagues/abc/market/status", nil)
	assertMarketStatus(t, w, http.StatusBadRequest)
}

func TestMarket_PlaceBid_NoAuthBeforeBody(t *testing.T) {
	h := &MarketHandler{}
	r := marketRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/1/market/bids", map[string]string{"bad": "body"})
	assertMarketStatus(t, w, http.StatusUnauthorized)
}
