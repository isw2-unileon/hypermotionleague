package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestJWTAuth_MissingHeader(t *testing.T) {
	w := doGet(setup(testSecret), "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidFormat_NoBearerPrefix(t *testing.T) {
	w := doGet(setup(testSecret), "Token abc123")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidFormat_BearerOnly(t *testing.T) {
	w := doGet(setup(testSecret), "Bearer")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	w := doGet(setup(testSecret), "Bearer this.is.not.a.valid.jwt")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	token, _ := auth.GenerateToken(42, "test@example.com", "secret-A")
	w := doGet(setup("secret-B"), "Bearer "+token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	token, _ := auth.GenerateToken(42, "test@example.com", testSecret)
	w := doGet(setup(testSecret), "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"user_id":42`) {
		t.Errorf("body %q missing user_id 42", w.Body.String())
	}
}

func TestJWTAuth_ValidToken_DifferentUser(t *testing.T) {
	token, _ := auth.GenerateToken(99, "otro@example.com", testSecret)
	w := doGet(setup(testSecret), "Bearer "+token)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"user_id":99`) {
		t.Errorf("body %q missing user_id 99", w.Body.String())
	}
}
