package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

// LeagueHandler handles HTTP requests for leagues
type LeagueHandler struct {
	repo repository.LeagueRepository
}

// NewLeagueHandler creates a new instance of the handler
func NewLeagueHandler(repo repository.LeagueRepository) *LeagueHandler {
	return &LeagueHandler{repo: repo}
}

func (h *LeagueHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	var req models.CreateLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de liga inválidos"})
		return
	}

	// Generate a random invite code
	code, err := generateInviteCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar código de invitación"})
		return
	}

	// Apply defaults
	maxMembers := req.MaxMembers
	if maxMembers == 0 {
		maxMembers = 10
	}
	budget := req.BudgetPerUser
	if budget == 0 {
		budget = 100000000 // 100M default
	}
	marketClose := req.MarketCloseTime
	if marketClose == "" {
		marketClose = "18:00:00"
	}

	league := &models.League{
		Name:            req.Name,
		InviteCode:      code,
		MaxMembers:      maxMembers,
		BudgetPerUser:   budget,
		MarketCloseTime: marketClose,
		CreatedBy:       userID,
	}

	if err := h.repo.Create(c.Request.Context(), league); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la liga"})
		return
	}

	// Add the creator as owner
	member := &models.LeagueMember{
		LeagueID: league.ID,
		UserID:   userID,
		Role:     models.RoleOwner,
		Budget:   budget,
	}
	if err := h.repo.AddMember(c.Request.Context(), member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Liga creada pero no se pudo añadir al creador"})
		return
	}

	c.JSON(http.StatusCreated, league)
}

// GetByID searches for a league by its ID and returns it
func (h *LeagueHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	league, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener liga"})
		return
	}
	if league == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Liga no encontrada"})
		return
	}

	c.JSON(http.StatusOK, league)
}

// GetByUserID returns the leagues that a user is part of, either as owner or member
func (h *LeagueHandler) GetByUserID(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	leagues, err := h.repo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener ligas"})
		return
	}

	if leagues == nil {
		leagues = []models.League{}
	}

	c.JSON(http.StatusOK, leagues)
}

// JoinLeague allows a user to join a league using an invitation code
func (h *LeagueHandler) JoinLeague(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	var req models.JoinLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código de invitación requerido"})
		return
	}

	league, err := h.repo.GetByInviteCode(c.Request.Context(), req.InviteCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al buscar liga"})
		return
	}
	if league == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Código de invitación no válido"})
		return
	}

	// Check if already a member
	existing, err := h.repo.GetMember(c.Request.Context(), league.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Ya eres miembro de esta liga"})
		return
	}

	// Check if league is full
	count, err := h.repo.CountMembers(c.Request.Context(), league.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}
	if count >= league.MaxMembers {
		c.JSON(http.StatusConflict, gin.H{"error": "La liga está llena"})
		return
	}

	member := &models.LeagueMember{
		LeagueID: league.ID,
		UserID:   userID,
		Role:     models.RoleMember,
		Budget:   league.BudgetPerUser,
	}

	if err := h.repo.AddMember(c.Request.Context(), member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo unir a la liga"})
		return
	}

	c.JSON(http.StatusOK, league)
}

// GetMembers returns the members of a league
func (h *LeagueHandler) GetMembers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	members, err := h.repo.GetMembers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener miembros"})
		return
	}

	if members == nil {
		members = []models.LeagueMember{}
	}

	c.JSON(http.StatusOK, members)
}

// Delete deletes a league (only the owner can do it)
func (h *LeagueHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	league, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno"})
		return
	}
	if league == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Liga no encontrada"})
		return
	}

	if league.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Solo el creador puede eliminar la liga"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo eliminar la liga"})
		return
	}

	c.Status(http.StatusNoContent)
}

func generateInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
