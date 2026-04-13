package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

// LineupHandler gestiona las peticiones HTTP relacionadas con alineaciones.
type LineupHandler struct {
	matchdayRepo repository.MatchdayRepository
	teamRepo     repository.TeamRepository
	leagueRepo   repository.LeagueRepository
}

// NewLineupHandler crea una nueva instancia de LineupHandler.
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

// getMatchdayByNumber es un helper que busca una jornada por su número dentro de una liga.
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

// GetLineup devuelve la alineación del usuario para una jornada concreta.
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

// SaveLineup crea o actualiza la alineación del usuario para una jornada.
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

	// Validar que hay exactamente 11 titulares
	starters := 0
	for _, p := range req.Players {
		if p.IsStarter {
			starters++
		}
	}
	if starters != 11 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La alineación debe tener exactamente 11 titulares"})
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

	// Obtener o crear el lineup
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

	// Guardar cada jugador en la alineación (upsert)
	for _, p := range req.Players {
		lp := &models.LineupPlayer{
			LineupID:  lineup.ID,
			PlayerID:  p.PlayerID,
			Position:  p.Position,
			IsStarter: p.IsStarter,
		}
		if err := h.matchdayRepo.UpsertLineupPlayer(c.Request.Context(), lp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar jugador en la alineación"})
			return
		}
	}

	// Devolver la alineación actualizada
	saved, err := h.matchdayRepo.GetLineup(c.Request.Context(), leagueID, userID, matchday.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al recuperar la alineación guardada"})
		return
	}

	c.JSON(http.StatusOK, saved)
}

// RemoveLineupPlayer elimina un jugador de la alineación del usuario.
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
