package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

// TeamHandler handles HTTP requests related to the user's device.
type TeamHandler struct {
	teamRepo   repository.TeamRepository
	leagueRepo repository.LeagueRepository
}

// NewTeamHandler creates a new instance of TeamHandler.
func NewTeamHandler(teamRepo repository.TeamRepository, leagueRepo repository.LeagueRepository) *TeamHandler {
	return &TeamHandler{teamRepo: teamRepo, leagueRepo: leagueRepo}
}

// GetUserTeam returns the complete team of the user in a league,
// including the list of players and their details.
// GET /api/v1/leagues/:id/team
func (h *TeamHandler) GetUserTeam(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	// Verify that the user belongs to the league
	member, err := h.leagueRepo.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar membresía"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return
	}

	team, err := h.teamRepo.GetUserTeam(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el equipo"})
		return
	}
	if team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Equipo no encontrado"})
		return
	}

	c.JSON(http.StatusOK, team)
}
