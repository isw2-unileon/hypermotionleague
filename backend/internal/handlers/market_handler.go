package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

// MarketHandler handles the logic for the market endpoints.
type MarketHandler struct {
	marketRepo *postgres.MarketRepo
	playerRepo *postgres.PlayerRepo
	teamRepo   *postgres.TeamRepo
	leagueRepo *postgres.LeagueRepo // ADDED: League repository for member validation
}

// NewMarketHandler creates a new instance by injecting the required repositories.
func NewMarketHandler(marketRepo *postgres.MarketRepo, playerRepo *postgres.PlayerRepo, teamRepo *postgres.TeamRepo, leagueRepo *postgres.LeagueRepo) *MarketHandler {
	return &MarketHandler{
		marketRepo: marketRepo,
		playerRepo: playerRepo,
		teamRepo:   teamRepo,
		leagueRepo: leagueRepo,
	}
}

// requireMember is a private helper that checks if a user is a member of the league
func (h *MarketHandler) requireMember(c *gin.Context, leagueID, userID int64) bool {
	member, err := h.leagueRepo.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar la membresía"})
		return false
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return false
	}
	return true
}

// 1. GetAvailablePlayers - Returns unsigned players in a specific league.
func (h *MarketHandler) GetAvailablePlayers(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	// Security Check: User must be a member of the league
	if !h.requireMember(c, leagueID, userID) {
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
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	// Security Check: User must be a member of the league
	if !h.requireMember(c, leagueID, userID) {
		return
	}

	listings, err := h.marketRepo.GetActiveListings(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve market listings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": listings})
}

// 3. PlaceBid - Validates and registers a new bid atomically.
func (h *MarketHandler) PlaceBid(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	if !h.requireMember(c, leagueID, userID) {
		return
	}

	var req models.PlaceBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx := c.Request.Context()

	// --- Security Listing Validations ---
	listing, err := h.marketRepo.GetListingByID(ctx, req.ListingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve listing"})
		return
	}
	if listing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Listing not found"})
		return
	}
	if listing.LeagueID != leagueID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cross-league IDOR detected: Listing does not belong to this league"})
		return
	}
	if listing.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusConflict, gin.H{"error": "Listing has expired"})
		return
	}

	// --- ATOMIC TRANSACTION ---
	// Create the bid model
	bid := &models.Bid{
		ListingID: req.ListingID,
		UserID:    userID,
		Amount:    req.Amount,
		Status:    "active",
	}

	// Execute transaction (Validates Max Bids and Commited Budget with a DB Lock)
	if err := h.marketRepo.PlaceBidTx(ctx, leagueID, bid); err != nil {
		if err.Error() == "MAX_BIDS_REACHED" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum limit of 5 active bids reached"})
			return
		}
		if err.Error() == "INSUFFICIENT_BUDGET" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient budget to place this bid (including committed funds)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place the bid securely"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Bid successfully placed", "data": bid})
}

// 4. GetUserBids - Returns the user's currently active bids.
func (h *MarketHandler) GetUserBids(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	// Security Check: User must be a member of the league
	if !h.requireMember(c, leagueID, userID) {
		return
	}

	bids, err := h.marketRepo.GetUserActiveBids(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve your bids"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": bids})
}

// 5. CancelBid - Cancels one of the user's active bids.
func (h *MarketHandler) CancelBid(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	// Security Check: User must be a member of the league
	if !h.requireMember(c, leagueID, userID) {
		return
	}

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
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid league ID"})
		return
	}

	// Security Check: User must be a member of the league
	if !h.requireMember(c, leagueID, userID) {
		return
	}

	status, err := h.marketRepo.GetMarketStatus(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve market status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": status})
}
