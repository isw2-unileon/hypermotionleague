package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

// MarketHandler handles the logic for the market endpoints.
type MarketHandler struct {
	marketRepo *postgres.MarketRepo
	playerRepo *postgres.PlayerRepo
	teamRepo   *postgres.TeamRepo // ADDED: Team repository for budget validation
}

// NewMarketHandler creates a new instance by injecting the required repositories.
func NewMarketHandler(marketRepo *postgres.MarketRepo, playerRepo *postgres.PlayerRepo, teamRepo *postgres.TeamRepo) *MarketHandler {
	return &MarketHandler{
		marketRepo: marketRepo,
		playerRepo: playerRepo,
		teamRepo:   teamRepo,
	}
}

// BidRequest defines the expected JSON structure from the frontend.
type BidRequest struct {
	ListingID int64   `json:"listing_id"`
	Amount    float64 `json:"amount"`
}

// 1. GetAvailablePlayers - Returns unsigned players in a specific league.
func (h *MarketHandler) GetAvailablePlayers(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	players, err := h.playerRepo.GetAvailableForLeague(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve available players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": players})
}

// 2. GetActiveListings - Returns the players currently listed on the market.
func (h *MarketHandler) GetActiveListings(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	listings, err := h.marketRepo.GetActiveListings(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve market listings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": listings})
}

// 3. PlaceBid - Validates and registers a new bid.
func (h *MarketHandler) PlaceBid(c *gin.Context) {
	// MOCK: Simulating user 1 until Auth logic is implemented by Dev 1.
	userID := int64(1)

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	var req BidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validation 1: Amount must be greater than 0
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bid amount must be greater than 0"})
		return
	}

	ctx := c.Request.Context()

	// Validation 2: Maximum 5 active bids per user
	activeBids, err := h.marketRepo.CountUserActiveBids(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check your active bids"})
		return
	}
	if activeBids >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum limit of 5 active bids reached"})
		return
	}

	// Validation 3: Sufficient budget (Using TeamRepo)
	userTeam, err := h.teamRepo.GetUserTeam(ctx, leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user budget"})
		return
	}
	if userTeam == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User team not found in this league"})
		return
	}
	if int(req.Amount) > userTeam.Budget {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient budget to place this bid"})
		return
	}

	// Create the bid model
	bid := &models.Bid{
		ListingID: req.ListingID,
		UserID:    userID,
		Amount:    int(req.Amount),
		Status:    "active",
	}

	// Save the bid to the database
	if err := h.marketRepo.PlaceBid(ctx, bid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place the bid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Bid successfully placed", "data": bid})
}

// 4. GetUserBids - Returns the user's currently active bids.
func (h *MarketHandler) GetUserBids(c *gin.Context) {
	userID := int64(1) // MOCK User

	bids, err := h.marketRepo.GetUserActiveBids(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve your bids"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": bids})
}

// 5. CancelBid - Cancels one of the user's active bids.
func (h *MarketHandler) CancelBid(c *gin.Context) {
	userID := int64(1) // MOCK User

	bidID, err := strconv.ParseInt(c.Param("bid_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bid ID"})
		return
	}

	if err := h.marketRepo.CancelBid(c.Request.Context(), bidID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not cancel bid (it may not exist or is already processed)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Bid successfully cancelled"})
}

// 6. GetMarketStatus - Returns the market status and closing time.
func (h *MarketHandler) GetMarketStatus(c *gin.Context) {
	userID := int64(1) // MOCK User
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	status, err := h.marketRepo.GetMarketStatus(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve market status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": status})
}
