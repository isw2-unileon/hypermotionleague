package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

type PlayerHandler struct {
	players   repository.PlayerRepository
	matchdays repository.MatchdayRepository
}

func NewPlayerHandler(players repository.PlayerRepository, matchdays repository.MatchdayRepository) *PlayerHandler {
	return &PlayerHandler{players: players, matchdays: matchdays}
}

// GET /api/v1/players?position=GK&team=Real+Madrid
func (h *PlayerHandler) List(c *gin.Context) {
	var position *models.PlayerPosition
	var teamName *string

	if pos := c.Query("position"); pos != "" {
		p := models.PlayerPosition(pos)
		position = &p
	}

	if team := c.Query("team"); team != "" {
		teamName = &team
	}

	players, err := h.players.List(c.Request.Context(), position, teamName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"players": players})
}

// GET /api/v1/players/:id
func (h *PlayerHandler) GetByID(c *gin.Context) {
	playerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	player, err := h.players.GetByID(c.Request.Context(), playerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	c.JSON(http.StatusOK, player)
}

// GET /api/v1/players/:id/points?matchday_id=3
func (h *PlayerHandler) GetPointsByMatchday(c *gin.Context) {
	playerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	matchdayIDStr := c.Query("matchday_id")
	if matchdayIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "matchday_id query param is required"})
		return
	}

	matchdayID, err := strconv.ParseInt(matchdayIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday_id"})
		return
	}

	points, err := h.players.GetPoints(c.Request.Context(), playerID, matchdayID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "points not found for this player and matchday"})
		return
	}

	c.JSON(http.StatusOK, points)
}
