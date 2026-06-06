package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/auth"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/middleware"
)

const testSecret = "test-secret-key-for-unit-tests"

func setup(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.JWTAuth(secret))
	r.GET("/protected", func(c *gin.Context) {
		userID := c.GetInt64(middleware.UserIDKey)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func doGet(r *gin.Engine, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── Missing header ───────────────────────────────────────────────────────────

func TestJWTAuth_MissingHeader(t *testing.T) {
	r := setup(testSecret)
	w := doGet(r, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ── Invalid format (no "Bearer " prefix) ─────────────────────────────────────

func TestJWTAuth_InvalidFormat_NoBearerPrefix(t *testing.T) {
	r := setup(testSecret)
	w := doGet(r, "Token abc123")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidFormat_BearerOnly(t *testing.T) {
	r := setup(testSecret)
	w := doGet(r, "Bearer")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidFormat_JustToken(t *testing.T) {
	r := setup(testSecret)
	w := doGet(r, "abc123")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ── Invalid / expired token ──────────────────────────────────────────────────

func TestJWTAuth_InvalidToken(t *testing.T) {
	r := setup(testSecret)
	w := doGet(r, "Bearer this.is.not.a.valid.jwt")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	// Generate token with one secret, validate with another
	token, err := auth.GenerateToken(42, "test@example.com", "secret-A")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	r := setup("secret-B")
	w := doGet(r, "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// ── Valid token ──────────────────────────────────────────────────────────────

func TestJWTAuth_ValidToken(t *testing.T) {
	token, err := auth.GenerateToken(42, "test@example.com", testSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	r := setup(testSecret)
	w := doGet(r, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	// Check that userID was injected into context
	body := w.Body.String()
	if !contains(body, `"user_id":42`) {
		t.Errorf("body %q does not contain user_id 42", body)
	}
}

func TestJWTAuth_ValidToken_DifferentUser(t *testing.T) {
	token, err := auth.GenerateToken(99, "otro@example.com", testSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	r := setup(testSecret)
	w := doGet(r, "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !contains(body, `"user_id":99`) {
		t.Errorf("body %q does not contain user_id 99", body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
