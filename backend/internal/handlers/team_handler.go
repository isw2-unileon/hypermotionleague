package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

type TeamHandler struct {
	teams    repository.TeamRepository
	teamRepo *postgres.TeamRepo // concrete repo for transactional methods (PayReleaseTx)
	leagues  repository.LeagueRepository
}

func NewTeamHandler(teams repository.TeamRepository, leagues repository.LeagueRepository, teamRepo *postgres.TeamRepo) *TeamHandler {
	return &TeamHandler{teams: teams, leagues: leagues, teamRepo: teamRepo}
}

// GET /api/v1/leagues/:id/team
func (h *TeamHandler) GetUserTeam(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	member, err := h.leagues.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify membership"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "caller is not a member of this league"})
		return
	}

	team, err := h.teams.GetUserTeam(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user team"})
		return
	}
	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// GET /api/v1/leagues/:id/users/:userId/team
func (h *TeamHandler) GetUserTeamInLeague(c *gin.Context) {
	callerID := c.GetInt64("userID")
	if callerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	callerMember, err := h.leagues.GetMember(c.Request.Context(), leagueID, callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify caller membership"})
		return
	}
	if callerMember == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "caller is not a member of this league"})
		return
	}

	targetMember, err := h.leagues.GetMember(c.Request.Context(), leagueID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify target membership"})
		return
	}
	if targetMember == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a member of this league"})
		return
	}

	team, err := h.teams.GetUserTeam(c.Request.Context(), leagueID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user team"})
		return
	}
	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

// POST /api/v1/leagues/:id/users/:userId/team/:playerId/buy
// Pay the release clause to buy a player from another user.
func (h *TeamHandler) BuyPlayer(c *gin.Context) {
	buyerID := c.GetInt64("userID")
	if buyerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	sellerID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	playerID, err := strconv.ParseInt(c.Param("playerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	if buyerID == sellerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot buy your own player"})
		return
	}

	// Verify buyer is a member
	buyerMember, err := h.leagues.GetMember(c.Request.Context(), leagueID, buyerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify membership"})
		return
	}
	if buyerMember == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this league"})
		return
	}

	clauseAmount, err := h.teamRepo.PayReleaseTx(c.Request.Context(), leagueID, buyerID, sellerID, playerID)
	if err != nil {
		switch err.Error() {
		case "PLAYER_NOT_OWNED":
			c.JSON(http.StatusNotFound, gin.H{"error": "player not found in seller's team"})
		case "INSUFFICIENT_BUDGET":
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient budget to pay release clause"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete purchase"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Player purchased via release clause",
		"amount":  clauseAmount,
	})
}
