package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func getWithOrigin(r *gin.Engine, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCORS_WildcardAllowsAll(t *testing.T) {
	w := getWithOrigin(corsRouter([]string{"*"}), "https://anything.com")
	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "*" {
		t.Errorf("Allow-Origin = %q, want *", h)
	}
}

func TestCORS_AllowedOrigin(t *testing.T) {
	w := getWithOrigin(corsRouter([]string{"https://myapp.vercel.app", "http://localhost:5173"}), "http://localhost:5173")
	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "http://localhost:5173" {
		t.Errorf("Allow-Origin = %q, want http://localhost:5173", h)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	w := getWithOrigin(corsRouter([]string{"https://myapp.vercel.app"}), "https://evil.com")
	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "" {
		t.Errorf("Allow-Origin = %q, want empty", h)
	}
}

func TestCORS_PreflightReturns204(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app"})
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://myapp.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestCORS_AuthorizationHeaderAllowed(t *testing.T) {
	r := corsRouter([]string{"https://myapp.vercel.app"})
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://myapp.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Errorf("Allow-Headers should contain Authorization")
	}
}

func TestCORS_MultipleOrigins(t *testing.T) {
	w := getWithOrigin(corsRouter([]string{"https://a.com", "https://b.com"}), "https://b.com")
	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "https://b.com" {
		t.Errorf("Allow-Origin = %q, want https://b.com", h)
	}
}
