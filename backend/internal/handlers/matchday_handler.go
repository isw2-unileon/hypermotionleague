package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

type MatchdayHandler struct {
	matchdays repository.MatchdayRepository
	leagues   repository.LeagueRepository
}

func NewMatchdayHandler(matchdays repository.MatchdayRepository, leagues repository.LeagueRepository) *MatchdayHandler {
	return &MatchdayHandler{matchdays: matchdays, leagues: leagues}
}

// requireMember reports whether the authenticated caller belongs to leagueID.
// On any negative outcome it writes the matching 401/403/500 response and
// returns false, so callers use: if !h.requireMember(c, leagueID) { return }.
// League-scoped reads (e.g. standings) must be gated by this so a member of one
// league cannot read another league's data by manipulating the :id.
func (h *MatchdayHandler) requireMember(c *gin.Context, leagueID int64) bool {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return false
	}
	member, err := h.leagues.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar membresía"})
		return false
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return false
	}
	return true
}

// GET /api/v1/leagues/:id/matchdays
// Matchdays are global; the :id (league) stays in the route for frontend
// compatibility but does not scope the lookup.
func (h *MatchdayHandler) GetByLeague(c *gin.Context) {
	if _, err := strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	matchdays, err := h.matchdays.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchdays"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matchdays": matchdays})
}

// GET /api/v1/leagues/:id/matchdays/current
func (h *MatchdayHandler) GetCurrent(c *gin.Context) {
	if _, err := strconv.ParseInt(c.Param("id"), 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	matchday, err := h.matchdays.GetCurrent(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no current matchday found"})
		return
	}

	c.JSON(http.StatusOK, matchday)
}

// GET /api/v1/leagues/:id/standings
func (h *MatchdayHandler) GetStandings(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	if !h.requireMember(c, leagueID) {
		return
	}

	standings, err := h.matchdays.GetStandings(c.Request.Context(), leagueID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch standings"})
		return
	}

	c.JSON(http.StatusOK, standings)
}

// GET /api/v1/leagues/:id/matchdays/:number/standings
func (h *MatchdayHandler) GetMatchdayStandings(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	if !h.requireMember(c, leagueID) {
		return
	}

	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday number"})
		return
	}

	// Matchdays are global; resolve the matchday by its app sequential number.
	matchday, err := h.matchdays.GetByNumber(c.Request.Context(), number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchday"})
		return
	}
	if matchday == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "matchday not found"})
		return
	}

	standings, err := h.matchdays.GetStandings(c.Request.Context(), leagueID, &matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchday standings"})
		return
	}

	c.JSON(http.StatusOK, standings)
}
