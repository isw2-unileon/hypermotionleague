package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func authRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func doAuthReq(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── Register tests ───────────────────────────────────────────────────────────

func TestAuth_Register_EmptyBody(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_MissingUsername(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"email": "test@example.com", "password": "12345678", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_MissingEmail(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "testuser", "password": "12345678", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_InvalidEmail(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "testuser", "email": "not-an-email", "password": "12345678", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_MissingPassword(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "testuser", "email": "test@example.com", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_PasswordTooShort(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "testuser", "email": "test@example.com", "password": "123", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_MissingDisplayName(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "testuser", "email": "test@example.com", "password": "12345678",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_UsernameTooShort(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", map[string]string{
		"username": "ab", "email": "test@example.com", "password": "12345678", "display_name": "Test",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Register_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/register", h.Register)
	w := doAuthReq(r, "POST", "/auth/register", "not json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ── Login tests ──────────────────────────────────────────────────────────────

func TestAuth_Login_EmptyBody(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/login", h.Login)
	w := doAuthReq(r, "POST", "/auth/login", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Login_MissingEmail(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/login", h.Login)
	w := doAuthReq(r, "POST", "/auth/login", map[string]string{"password": "12345678"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Login_MissingPassword(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/login", h.Login)
	w := doAuthReq(r, "POST", "/auth/login", map[string]string{"email": "test@example.com"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Login_InvalidEmail(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/login", h.Login)
	w := doAuthReq(r, "POST", "/auth/login", map[string]string{"email": "bad", "password": "12345678"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuth_Login_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, "secret")
	r := authRouter()
	r.POST("/auth/login", h.Login)
	w := doAuthReq(r, "POST", "/auth/login", "not json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
