package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

type MatchdayHandler struct {
	matchdays repository.MatchdayRepository
}

func NewMatchdayHandler(matchdays repository.MatchdayRepository) *MatchdayHandler {
	return &MatchdayHandler{matchdays: matchdays}
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
