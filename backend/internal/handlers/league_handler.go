package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

// LeagueHandler gestiona las peticiones HTTP para las ligas
type LeagueHandler struct {
	repo repository.LeagueRepository
}

// NewLeagueHandler crea una nueva instancia del handler
func NewLeagueHandler(repo repository.LeagueRepository) *LeagueHandler {
	return &LeagueHandler{repo: repo}
}

func (h *LeagueHandler) Create(c *gin.Context) {
	var req models.CreateLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de liga inválidos"})
		return
	}

	league := &models.League{
		Name:          req.Name,
		MaxMembers:    req.MaxMembers,
		BudgetPerUser: req.BudgetPerUser,
	}

	if err := h.repo.Create(c.Request.Context(), league); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la liga"})
		return
	}

	c.JSON(http.StatusCreated, league)
}

// GetByID busca una liga por su ID
func (h *LeagueHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	league, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Liga no encontrada"})
		return
	}

	c.JSON(http.StatusOK, league)
}

// GetByUserID devuelve las ligas del usuario autenticado
func (h *LeagueHandler) GetByUserID(c *gin.Context) {
	// Aquí usamos tu instrucción estrella:
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

// AddMember permite a un usuario unirse a una liga
func (h *LeagueHandler) AddMember(c *gin.Context) {
	var member models.LeagueMember
	if err := c.ShouldBindJSON(&member); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de miembro inválidos"})
		return
	}

	if err := h.repo.AddMember(c.Request.Context(), &member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo unir a la liga"})
		return
	}

	c.JSON(http.StatusCreated, member)
}
