package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

// LineupHandler handles HTTP requests related to lineups.
type LineupHandler struct {
	matchdayRepo repository.MatchdayRepository
	teamRepo     repository.TeamRepository
	leagueRepo   repository.LeagueRepository
}

// NewLineupHandler creates a new instance of LineupHandler.
func NewLineupHandler(
	matchdayRepo repository.MatchdayRepository,
	teamRepo repository.TeamRepository,
	leagueRepo repository.LeagueRepository,
) *LineupHandler {
	return &LineupHandler{
		matchdayRepo: matchdayRepo,
		teamRepo:     teamRepo,
		leagueRepo:   leagueRepo,
	}
}

// getMatchdayByNumber  is a helper that searches for a matchday by its number within a league.
func (h *LineupHandler) getMatchdayByNumber(c *gin.Context, leagueID int64, number int) (*models.Matchday, bool) {
	matchdays, err := h.matchdayRepo.GetByLeague(c.Request.Context(), leagueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener jornadas"})
		return nil, false
	}
	for _, m := range matchdays {
		if m.Number == number {
			return &m, true
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Jornada no encontrada"})
	return nil, false
}

// GetLineup returns the lineup of the authenticated user for a specific matchday.
// GET /api/v1/leagues/:id/matchdays/:number/lineup
func (h *LineupHandler) GetLineup(c *gin.Context) {
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

	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número de jornada inválido"})
		return
	}

	member, err := h.leagueRepo.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar membresía"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return
	}

	matchday, ok := h.getMatchdayByNumber(c, leagueID, number)
	if !ok {
		return
	}

	lineup, err := h.matchdayRepo.GetLineup(c.Request.Context(), leagueID, userID, matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la alineación"})
		return
	}
	if lineup == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alineación no encontrada para esta jornada"})
		return
	}

	c.JSON(http.StatusOK, lineup)
}

// SaveLineup creates or updates the user's lineup for a matchday.
// PUT /api/v1/leagues/:id/matchdays/:number/lineup
func (h *LineupHandler) SaveLineup(c *gin.Context) {
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

	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número de jornada inválido"})
		return
	}

	var req models.CreateLineupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de alineación inválidos"})
		return
	}

	member, err := h.leagueRepo.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar membresía"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return
	}

	matchday, ok := h.getMatchdayByNumber(c, leagueID, number)
	if !ok {
		return
	}

	if !matchday.StartDate.After(time.Now()) {
		c.JSON(http.StatusConflict, gin.H{"error": "matchday already started"})
		return
	}

	// validate formation: exactly 11 starters with 1 GK, 3-5 DEF, 3-5 MID, 1-3 FWD
	byPos := map[models.PlayerPosition]int{}
	starters := 0
	for _, p := range req.Players {
		if p.IsStarter {
			starters++
			byPos[p.Position]++
		}
	}
	if starters != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La alineación debe tener exactamente 11 titulares"})
		return
	}
	gk, def, mid, fwd := byPos[models.PositionGK], byPos[models.PositionDEF], byPos[models.PositionMID], byPos[models.PositionFWD]
	if gk != 1 || def < 3 || def > 5 || mid < 3 || mid > 5 || fwd < 1 || fwd > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formación inválida: se requiere 1 GK, 3-5 DEF, 3-5 MID, 1-3 FWD"})
		return
	}

	// Validar que todos los jugadores pertenecen al equipo del usuario
	for _, p := range req.Players {
		owned, err := h.teamRepo.HasPlayer(c.Request.Context(), leagueID, userID, p.PlayerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar jugadores"})
			return
		}
		if !owned {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Jugador no pertenece a tu equipo"})
			return
		}
	}

	// Obtain or create lineup for this matchday
	lineup, err := h.matchdayRepo.GetLineup(c.Request.Context(), leagueID, userID, matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la alineación"})
		return
	}
	if lineup == nil {
		newLineup := &models.Lineup{
			LeagueID:   leagueID,
			UserID:     userID,
			MatchdayID: matchday.ID,
		}
		if err := h.matchdayRepo.CreateLineup(c.Request.Context(), newLineup); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear la alineación"})
			return
		}
		lineup = &models.LineupWithPlayers{Lineup: *newLineup}
	}

	players := make([]models.LineupPlayer, len(req.Players))
	for i, p := range req.Players {
		players[i] = models.LineupPlayer{
			LineupID:  lineup.ID,
			PlayerID:  p.PlayerID,
			Position:  p.Position,
			IsStarter: p.IsStarter,
		}
	}
	if err := h.matchdayRepo.ReplaceLineupPlayers(c.Request.Context(), lineup.ID, players); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar la alineación"})
		return
	}

	// Devolver la alineación actualizada
	saved, err := h.matchdayRepo.GetLineup(c.Request.Context(), leagueID, userID, matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al recuperar la alineación guardada"})
		return
	}

	c.JSON(http.StatusOK, saved)
}

// RemoveLineupPlayer deletes a player from the user's lineup for a matchday.
// DELETE /api/v1/leagues/:id/matchdays/:number/lineup/players/:player_id
func (h *LineupHandler) RemoveLineupPlayer(c *gin.Context) {
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

	number, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Número de jornada inválido"})
		return
	}

	playerID, err := strconv.ParseInt(c.Param("player_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de jugador inválido"})
		return
	}

	member, err := h.leagueRepo.GetMember(c.Request.Context(), leagueID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al verificar membresía"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres miembro de esta liga"})
		return
	}

	matchday, ok := h.getMatchdayByNumber(c, leagueID, number)
	if !ok {
		return
	}

	lineup, err := h.matchdayRepo.GetLineup(c.Request.Context(), leagueID, userID, matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener la alineación"})
		return
	}
	if lineup == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No hay alineación para esta jornada"})
		return
	}

	if err := h.matchdayRepo.RemoveLineupPlayer(c.Request.Context(), lineup.ID, playerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar el jugador de la alineación"})
		return
	}

	c.Status(http.StatusNoContent)
}
