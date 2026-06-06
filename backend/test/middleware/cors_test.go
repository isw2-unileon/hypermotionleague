package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/middleware"
)

func corsRouter(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS(origins))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func preflight(r *gin.Engine, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getWithOrigin(r *gin.Engine, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── Wildcard origin "*" ──────────────────────────────────────────────────────

func TestCORS_WildcardAllowsAll(t *testing.T) {
	r := corsRouter([]string{"*"})
	w := getWithOrigin(r, "https://anything.example.com")
	header := w.Header().Get("Access-Control-Allow-Origin")
	if header != "*" {
		t.Errorf("Allow-Origin = %q, want *", header)
	}
}

// ── Specific origin allowed ──────────────────────────────────────────────────

func TestCORS_AllowedOrigin(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app", "http://localhost:5173"})
	w := getWithOrigin(r, "http://localhost:5173")
	header := w.Header().Get("Access-Control-Allow-Origin")
	if header != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", header)
	}
}

// ── Origin not in list ───────────────────────────────────────────────────────

func TestCORS_DisallowedOrigin(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app"})
	w := getWithOrigin(r, "https://evil.com")
	header := w.Header().Get("Access-Control-Allow-Origin")
	if header != "" {
		t.Errorf("Allow-Origin = %q, want empty (blocked)", header)
	}
}

// ── Preflight returns 204 ────────────────────────────────────────────────────

func TestCORS_PreflightReturns204(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app"})
	w := preflight(r, "https://myapp.vercel.app")
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

// ── Authorization header is allowed ──────────────────────────────────────────

func TestCORS_AuthorizationHeaderAllowed(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app"})
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://myapp.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if !containsStr(allowHeaders, "Authorization") {
		t.Errorf("Allow-Headers = %q, should contain Authorization", allowHeaders)
	}
}

// ── Multiple origins: only matching one is echoed ────────────────────────────

func TestCORS_MultipleOrigins_OnlyMatchingEchoed(t *testing.T) {
	r := corsRouter([]string{"https://a.com", "https://b.com", "https://c.com"})
	w := getWithOrigin(r, "https://b.com")
	header := w.Header().Get("Access-Control-Allow-Origin")
	if header != "https://b.com" {
		t.Errorf("Allow-Origin = %q, want https://b.com", header)
	}
}
