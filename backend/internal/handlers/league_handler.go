package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	mrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/market"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository"
)

var leagueLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// matchdayLister is the slice of the matchday repo the league handler needs to
// evaluate the market window for a league (its matchdays). Satisfied by
// *postgres.MatchdayRepo.
type matchdayLister interface {
	GetByLeague(ctx context.Context, leagueID int64) ([]models.Matchday, error)
}

// LeagueHandler handles HTTP requests for leagues
type LeagueHandler struct {
	repo  repository.LeagueRepository
	teams repository.TeamRepository

	// Market-seeding collaborators. When all are non-nil, creating a league
	// while the market is open immediately stocks its market (see seedMarketIfOpen).
	// They are optional so the handler can be built without them (e.g. in tests).
	marketStore market.ListingStore
	players     market.FreeAgentSource
	matchdays   matchdayLister
	loc         *time.Location
}

// NewLeagueHandler creates a new instance of the handler. The market-seeding
// collaborators (marketStore, players, matchdays) may be nil to disable the
// on-create market seeding. The Europe/Madrid location is loaded once (falling
// back to UTC if the embedded tz data is unavailable).
func NewLeagueHandler(
	repo repository.LeagueRepository,
	teams repository.TeamRepository,
	marketStore market.ListingStore,
	players market.FreeAgentSource,
	matchdays matchdayLister,
) *LeagueHandler {
	loc, err := market.Location()
	if err != nil {
		loc = time.UTC
	}
	return &LeagueHandler{
		repo:        repo,
		teams:       teams,
		marketStore: marketStore,
		players:     players,
		matchdays:   matchdays,
		loc:         loc,
	}
}

// seedMarketIfOpen stocks a freshly created league's market if the trading
// window is open right now (Sprint 3 1.A rule). A new league usually has no
// matchdays, so "open" then depends only on the 19:00–00:00 Europe/Madrid
// window. It is best-effort: it never returns an error and must never block or
// undo league creation — if the market is closed or seeding fails, the daily
// cron will stock the league later.
func (h *LeagueHandler) seedMarketIfOpen(ctx context.Context, leagueID int64) {
	if h.marketStore == nil || h.players == nil || h.matchdays == nil {
		return // market seeding not wired (e.g. in unit tests)
	}

	matchdays, err := h.matchdays.GetByLeague(ctx, leagueID)
	if err != nil {
		leagueLogger.Warn("market seed: could not load matchdays; skipping",
			"league_id", leagueID, "error", err)
		return
	}

	if w := market.ComputeWindow(time.Now(), h.loc, matchdays); !w.IsOpen {
		return // market closed now → the daily cron will populate it
	}

	// #nosec G404 -- el mercado no necesita aleatoriedad criptográfica.
	rng := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	created, _, err := market.RefreshLeague(
		ctx, h.marketStore, h.players, leagueID, market.DefaultTarget, time.Now().Add(market.ListingTTL), rng)
	if err != nil {
		leagueLogger.Warn("market seed failed; league created without market (cron will fill it)",
			"league_id", leagueID, "error", err)
		return
	}
	leagueLogger.Info("seeded new league market", "league_id", leagueID, "created", created)
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

	// Draft the owner's initial squad (best-effort: an empty player pool
	// must not prevent league creation — e.g. in test/CI environments).
	_ = h.teams.DraftInitialSquad(c.Request.Context(), league.ID, userID)

	// If the market is open right now, stock the new league's market immediately
	// so it is not empty until the daily cron. Best-effort: never blocks or
	// undoes league creation (see seedMarketIfOpen).
	h.seedMarketIfOpen(c.Request.Context(), league.ID)

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

	// Draft the new member's initial squad (best-effort).
	_ = h.teams.DraftInitialSquad(c.Request.Context(), league.ID, userID)

	c.JSON(http.StatusOK, league)
}

// GetMembers returns the members of a league
func (h *LeagueHandler) GetMembers(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de liga inválido"})
		return
	}

	members, err := h.repo.GetMembersWithUsers(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener miembros"})
		return
	}

	if members == nil {
		members = []models.LeagueMemberWithUser{}
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
