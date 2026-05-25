package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

type TeamHandler struct {
	teams   repository.TeamRepository
	leagues repository.LeagueRepository
}

func NewTeamHandler(teams repository.TeamRepository, leagues repository.LeagueRepository) *TeamHandler {
	return &TeamHandler{teams: teams, leagues: leagues}
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
