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
func (h *MatchdayHandler) GetByLeague(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	matchdays, err := h.matchdays.GetByLeague(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchdays"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matchdays": matchdays})
}

// GET /api/v1/leagues/:id/matchdays/current
func (h *MatchdayHandler) GetCurrent(c *gin.Context) {
	leagueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid league id"})
		return
	}

	matchday, err := h.matchdays.GetCurrent(c.Request.Context(), leagueID)
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

	number, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday number"})
		return
	}

	// Find the matchday by number within this league
	matchdays, err := h.matchdays.GetByLeague(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchdays"})
		return
	}

	var matchdayID int64
	found := false
	for _, md := range matchdays {
		if int64(md.Number) == number {
			matchdayID = md.ID
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "matchday not found"})
		return
	}

	standings, err := h.matchdays.GetStandings(c.Request.Context(), leagueID, &matchdayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matchday standings"})
		return
	}

	c.JSON(http.StatusOK, standings)
}
