package handlers_test

import (
	"net/http"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
	h "github.com/isw2-unileon/proyect-scaffolding/backend/test/helpers"
)

const testJWTSecret = "test-secret-for-auth-handler"

// ── Register tests ───────────────────────────────────────────────────────────

func TestRegister_EmptyBody(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	w := h.DoReq(r, "POST", "/auth/register", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_MissingUsername(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"email":        "test@example.com",
		"password":     "12345678",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_MissingEmail(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username":     "testuser",
		"password":     "12345678",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_InvalidEmail(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username":     "testuser",
		"email":        "not-an-email",
		"password":     "12345678",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_MissingPassword(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username":     "testuser",
		"email":        "test@example.com",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username":     "testuser",
		"email":        "test@example.com",
		"password":     "123",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_MissingDisplayName(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "12345678",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_UsernameTooShort(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	body := map[string]string{
		"username":     "ab",
		"email":        "test@example.com",
		"password":     "12345678",
		"display_name": "Test User",
	}
	w := h.DoReq(r, "POST", "/auth/register", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestRegister_InvalidJSON(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/register", ah.Register)

	w := h.DoReq(r, "POST", "/auth/register", "not json")
	h.AssertStatus(t, w, http.StatusBadRequest)
}

// ── Login tests ──────────────────────────────────────────────────────────────

func TestLogin_EmptyBody(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/login", ah.Login)

	w := h.DoReq(r, "POST", "/auth/login", nil)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_MissingEmail(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/login", ah.Login)

	body := map[string]string{
		"password": "12345678",
	}
	w := h.DoReq(r, "POST", "/auth/login", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_MissingPassword(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/login", ah.Login)

	body := map[string]string{
		"email": "test@example.com",
	}
	w := h.DoReq(r, "POST", "/auth/login", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_InvalidEmail(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/login", ah.Login)

	body := map[string]string{
		"email":    "not-valid",
		"password": "12345678",
	}
	w := h.DoReq(r, "POST", "/auth/login", body)
	h.AssertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_InvalidJSON(t *testing.T) {
	ah := handlers.NewAuthHandler(nil, testJWTSecret)
	r := h.NewRouter(0)
	r.POST("/auth/login", ah.Login)

	w := h.DoReq(r, "POST", "/auth/login", "not json")
	h.AssertStatus(t, w, http.StatusBadRequest)
}
