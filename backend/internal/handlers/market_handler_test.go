package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func marketTestRouter(userID int64) *gin.Engine {
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

// ── Auth guard tests ─────────────────────────────────────────────────────────

func TestMarket_GetAvailablePlayers_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)
	w := doMarketReq(r, "GET", "/leagues/1/market/players", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarket_GetActiveListings_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)
	w := doMarketReq(r, "GET", "/leagues/1/market/listings", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarket_PlaceBid_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/1/market/bids", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarket_GetUserBids_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.GET("/leagues/:id/market/bids", h.GetUserBids)
	w := doMarketReq(r, "GET", "/leagues/1/market/bids", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarket_CancelBid_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)
	w := doMarketReq(r, "DELETE", "/leagues/1/market/bids/1", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMarket_GetMarketStatus_NoAuth(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)
	w := doMarketReq(r, "GET", "/leagues/1/market/status", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ── Invalid league ID tests ──────────────────────────────────────────────────

func TestMarket_GetAvailablePlayers_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(1)
	r.GET("/leagues/:id/market/players", h.GetAvailablePlayers)
	w := doMarketReq(r, "GET", "/leagues/abc/market/players", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarket_GetActiveListings_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(1)
	r.GET("/leagues/:id/market/listings", h.GetActiveListings)
	w := doMarketReq(r, "GET", "/leagues/xyz/market/listings", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarket_PlaceBid_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(1)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/abc/market/bids", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarket_GetMarketStatus_InvalidLeagueID(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(1)
	r.GET("/leagues/:id/market/status", h.GetMarketStatus)
	w := doMarketReq(r, "GET", "/leagues/abc/market/status", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMarket_PlaceBid_NoAuthBeforeBody(t *testing.T) {
	h := &MarketHandler{}
	r := marketTestRouter(0)
	r.POST("/leagues/:id/market/bids", h.PlaceBid)
	w := doMarketReq(r, "POST", "/leagues/1/market/bids", map[string]string{"bad": "body"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
