package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/auth"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo  *postgres.UserRepo
	jwtSecret string
}

func NewAuthHandler(userRepo *postgres.UserRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

// Register godoc
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// verify if email already exists
	existing, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		slog.Error("register: failed to check existing email", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// Hashing the password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("register: failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	hashStr := string(hash)
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: &hashStr,
		DisplayName:  req.DisplayName,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		slog.Error("register: failed to create user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		slog.Error("register: failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: *user})
}

// Login godoc
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		slog.Error("login: failed to query user", "error", err, "email", req.Email)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// OAuth users don't have a password — they can't use email/password login
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		slog.Error("login: failed to generate token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: *user})
}
