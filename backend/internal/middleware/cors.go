package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS builds the CORS middleware from an explicit allow-list of origins.
//
// It must be registered on the root router BEFORE any routes so that browser
// preflight OPTIONS requests are answered (with 204 + the CORS headers) by this
// middleware instead of falling through to Gin's router and returning 404.
//
// Wildcard subdomains (e.g. "https://*.vercel.app") are supported via
// AllowWildcard so Vercel preview deployments are accepted. Credentials are
// disabled: the app authenticates with a Bearer token in the Authorization
// header, not cookies, which also keeps the wildcard origin conflict-free
// (the CORS spec forbids credentials together with a wildcard origin).
//
// If the list contains "*", all origins are allowed (legacy behaviour).
func CORS(allowedOrigins []string) gin.HandlerFunc {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowWildcard:    true,
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	for _, o := range allowedOrigins {
		if o == "*" {
			cfg.AllowAllOrigins = true
			break
		}
	}
	if !cfg.AllowAllOrigins {
		cfg.AllowOrigins = allowedOrigins
	}

	return cors.New(cfg)
}
