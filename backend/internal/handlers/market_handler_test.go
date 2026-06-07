package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

// ── fake repositories for business-logic tests ───────────────────────────────

type fakeLeagueRepoMarket struct {
	repository.LeagueRepository
	getMember func(context.Context, int64, int64) (*models.LeagueMember, error)
}

func (f *fakeLeagueRepoMarket) GetMember(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error) {
	return f.getMember(ctx, leagueID, userID)
}

type fakeMarketRepoForBiz struct {
	repository.MarketRepository
	getListingByID func(context.Context, int64) (*models.MarketListingWithDetails, error)
	placeBidTx     func(context.Context, int64, *models.Bid) error
	cancelBid      func(context.Context, int64, int64) error
	insertEvent    func(context.Context, int64, int64, string, string, int, string) error
}

func (f *fakeMarketRepoForBiz) GetListingByID(ctx context.Context, id int64) (*models.MarketListingWithDetails, error) {
	return f.getListingByID(ctx, id)
}
func (f *fakeMarketRepoForBiz) PlaceBidTx(ctx context.Context, leagueID int64, bid *models.Bid) error {
	return f.placeBidTx(ctx, leagueID, bid)
}
func (f *fakeMarketRepoForBiz) CancelBid(ctx context.Context, bidID, userID int64) error {
	return f.cancelBid(ctx, bidID, userID)
}
func (f *fakeMarketRepoForBiz) InsertEvent(ctx context.Context, leagueID, userID int64, eventType, playerName string, amount int, details string) error {
	if f.insertEvent != nil {
		return f.insertEvent(ctx, leagueID, userID, eventType, playerName, amount, details)
	}
	return nil
}

type fakeMatchdayRepoMarket struct {
	repository.MatchdayRepository
	getAll func(context.Context) ([]models.Matchday, error)
}

func (f *fakeMatchdayRepoMarket) GetAll(ctx context.Context) ([]models.Matchday, error) {
	return f.getAll(ctx)
}

type fakeTeamRepoMarket struct {
	payClauseTx func(context.Context, int64, int64, int64) (*models.ClauseResult, error)
}

func (f *fakeTeamRepoMarket) PayClauseTx(ctx context.Context, leagueID, buyerID, playerID int64) (*models.ClauseResult, error) {
	return f.payClauseTx(ctx, leagueID, buyerID, playerID)
}

// validMemberLeague returns a league repo that always approves membership.
func validMemberLeague() *fakeLeagueRepoMarket {
	return &fakeLeagueRepoMarket{
		getMember: func(_ context.Context, leagueID, userID int64) (*models.LeagueMember, error) {
			return &models.LeagueMember{LeagueID: leagueID, UserID: userID}, nil
		},
	}
}

// newMarketHandlerForTest builds a MarketHandler with the given fakes.
func newMarketHandlerForTest(
	league repository.LeagueRepository,
	mkt repository.MarketRepository,
	matchday repository.MatchdayRepository,
	team clausePayer,
) *MarketHandler {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		loc = time.UTC
	}
	return &MarketHandler{
		leagueRepo:   league,
		marketRepo:   mkt,
		matchdayRepo: matchday,
		teamRepo:     team,
		loc:          loc,
	}
}

// activeMatchdays returns a list with one in-progress matchday, making the
// market window closed regardless of the current time-of-day.
func activeMatchdays() []models.Matchday {
	return []models.Matchday{{
		StartDate: time.Now().Add(-1 * time.Hour),
		EndDate:   time.Now().Add(24 * time.Hour),
	}}
}

// ── PlaceBid business-logic tests ────────────────────────────────────────────

func TestMarket_PlaceBid_BusinessLogic(t *testing.T) {
	validListing := &models.MarketListingWithDetails{
		MarketListing: models.MarketListing{
			ID:        10,
			LeagueID:  1,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	cases := []struct {
		name       string
		marketRepo *fakeMarketRepoForBiz
		body       any
		wantStatus int
	}{
		{
			name: "listing_not_found",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					return nil, nil
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "cross_league_idor",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					l := *validListing
					l.LeagueID = 999
					return &l, nil
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "expired_listing",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					l := *validListing
					l.ExpiresAt = time.Now().Add(-1 * time.Hour)
					return &l, nil
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusConflict,
		},
		{
			name: "max_bids_reached",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					return validListing, nil
				},
				placeBidTx: func(_ context.Context, _ int64, _ *models.Bid) error {
					return errors.New("MAX_BIDS_REACHED")
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "insufficient_budget",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					return validListing, nil
				},
				placeBidTx: func(_ context.Context, _ int64, _ *models.Bid) error {
					return errors.New("INSUFFICIENT_BUDGET")
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "success",
			marketRepo: &fakeMarketRepoForBiz{
				getListingByID: func(_ context.Context, _ int64) (*models.MarketListingWithDetails, error) {
					return validListing, nil
				},
				placeBidTx: func(_ context.Context, _ int64, _ *models.Bid) error {
					return nil
				},
			},
			body:       map[string]any{"listing_id": 10, "amount": 100},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newMarketHandlerForTest(validMemberLeague(), tc.marketRepo, nil, nil)
			r := marketTestRouter(1)
			r.POST("/leagues/:id/market/bids", h.PlaceBid)
			w := doMarketReq(r, "POST", "/leagues/1/market/bids", tc.body)
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

// ── CancelBid business-logic tests ───────────────────────────────────────────

func TestMarket_CancelBid_BusinessLogic(t *testing.T) {
	cases := []struct {
		name       string
		cancelBid  func(context.Context, int64, int64) error
		wantStatus int
	}{
		{
			name: "bid_not_found_or_processed",
			cancelBid: func(_ context.Context, _, _ int64) error {
				return errors.New("bid not found")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "success",
			cancelBid: func(_ context.Context, _, _ int64) error {
				return nil
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mkt := &fakeMarketRepoForBiz{cancelBid: tc.cancelBid}
			h := newMarketHandlerForTest(validMemberLeague(), mkt, nil, nil)
			r := marketTestRouter(1)
			r.DELETE("/leagues/:id/market/bids/:bid_id", h.CancelBid)
			w := doMarketReq(r, "DELETE", "/leagues/1/market/bids/5", nil)
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

// ── PayClause business-logic tests ───────────────────────────────────────────

func TestMarket_PayClause_BusinessLogic(t *testing.T) {
	noMatchdays := func(_ context.Context) ([]models.Matchday, error) { return nil, nil }
	closedMatchdays := func(_ context.Context) ([]models.Matchday, error) {
		return activeMatchdays(), nil
	}

	cases := []struct {
		name        string
		getAll      func(context.Context) ([]models.Matchday, error)
		payClauseTx func(context.Context, int64, int64, int64) (*models.ClauseResult, error)
		wantStatus  int
	}{
		{
			name:        "market_closed",
			getAll:      closedMatchdays,
			payClauseTx: nil,
			wantStatus:  http.StatusConflict,
		},
		{
			name:   "player_is_free_agent",
			getAll: noMatchdays,
			payClauseTx: func(_ context.Context, _, _, _ int64) (*models.ClauseResult, error) {
				return nil, postgres.ErrClausePlayerFree
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "self_purchase",
			getAll: noMatchdays,
			payClauseTx: func(_ context.Context, _, _, _ int64) (*models.ClauseResult, error) {
				return nil, postgres.ErrClauseSelfPurchase
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "insufficient_budget",
			getAll: noMatchdays,
			payClauseTx: func(_ context.Context, _, _, _ int64) (*models.ClauseResult, error) {
				return nil, postgres.ErrClauseInsufficient
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "buyer_not_member",
			getAll: noMatchdays,
			payClauseTx: func(_ context.Context, _, _, _ int64) (*models.ClauseResult, error) {
				return nil, postgres.ErrClauseBuyerNotMember
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "success",
			getAll: noMatchdays,
			payClauseTx: func(_ context.Context, _, _, _ int64) (*models.ClauseResult, error) {
				return &models.ClauseResult{}, nil
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matchdayRepo := &fakeMatchdayRepoMarket{getAll: tc.getAll}
			var team clausePayer
			if tc.payClauseTx != nil {
				team = &fakeTeamRepoMarket{payClauseTx: tc.payClauseTx}
			}
			h := newMarketHandlerForTest(validMemberLeague(), nil, matchdayRepo, team)
			r := marketTestRouter(1)
			r.POST("/leagues/:id/market/clause/:player_id", h.PayClause)
			w := doMarketReq(r, "POST", "/leagues/1/market/clause/42", nil)
			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

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
