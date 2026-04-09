package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/auth"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

type OAuthHandler struct {
	userRepo    *postgres.UserRepo
	jwtSecret   string
	supabaseURL string
	supabaseKey string
}

func NewOAuthHandler(userRepo *postgres.UserRepo, jwtSecret, supabaseURL, supabaseKey string) *OAuthHandler {
	return &OAuthHandler{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		supabaseURL: supabaseURL,
		supabaseKey: supabaseKey,
	}
}

// supabaseUser represents the user info returned by Supabase's /auth/v1/user endpoint.
type supabaseUser struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	AppMetadata      map[string]interface{} `json:"app_metadata"`
	UserMetadata     map[string]interface{} `json:"user_metadata"`
	Identities       []supabaseIdentity     `json:"identities"`
}

type supabaseIdentity struct {
	Provider string `json:"provider"`
}

// Callback handles POST /api/v1/auth/oauth
// The frontend sends the Supabase access_token after a successful OAuth flow.
// This endpoint verifies it with Supabase, then finds or creates the user in our DB.
func (h *OAuthHandler) Callback(c *gin.Context) {
	var req models.OAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the token with Supabase by calling their /auth/v1/user endpoint
	supaUser, err := h.getSupabaseUser(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid oauth token"})
		return
	}

	// Determine provider from Supabase identities
	provider := "oauth"
	if len(supaUser.Identities) > 0 {
		provider = supaUser.Identities[0].Provider
	}

	// Check if user already exists in our DB
	user, err := h.userRepo.GetByEmail(c.Request.Context(), supaUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if user == nil {
		// Extract display name from Supabase metadata
		displayName := extractName(supaUser.UserMetadata)
		username := generateUsername(supaUser.Email)
		avatarURL := extractString(supaUser.UserMetadata, "avatar_url")

		user = &models.User{
			Username:     username,
			Email:        supaUser.Email,
			PasswordHash: "", // OAuth users don't have a password
			DisplayName:  displayName,
			AvatarURL:    avatarURL,
			AuthProvider: provider,
		}

		if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: *user})
}

// getSupabaseUser verifies the access token by calling Supabase Auth API.
func (h *OAuthHandler) getSupabaseUser(accessToken string) (*supabaseUser, error) {
	req, err := http.NewRequest("GET", h.supabaseURL+"/auth/v1/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", h.supabaseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supabase returned status %d", resp.StatusCode)
	}

	var user supabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	if user.Email == "" {
		return nil, fmt.Errorf("no email in supabase user")
	}

	return &user, nil
}

// extractName gets a display name from Supabase user metadata.
func extractName(meta map[string]interface{}) string {
	for _, key := range []string{"full_name", "name", "preferred_username"} {
		if v, ok := meta[key].(string); ok && v != "" {
			return v
		}
	}
	return "Manager"
}

// extractString gets a string value from metadata.
func extractString(meta map[string]interface{}, key string) *string {
	if v, ok := meta[key].(string); ok && v != "" {
		return &v
	}
	return nil
}

// generateUsername creates a username from an email address.
func generateUsername(email string) string {
	parts := strings.Split(email, "@")
	return parts[0]
}
