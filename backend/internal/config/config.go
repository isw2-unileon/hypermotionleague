// Package config handles application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// defaultCORSOrigins is the comma-separated allow-list used when neither
// CORS_ALLOW_ORIGINS nor CORS_ALLOW_ORIGIN is set: the Vercel production
// domain, Vercel preview deployments (wildcard), and local dev.
const defaultCORSOrigins = "https://hypermotionleague.vercel.app,https://*.vercel.app,http://localhost:5173,http://localhost:3000"

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port             string
	GinMode          string
	CORSAllowOrigin  string   // deprecated single-origin value, kept for compatibility
	CORSAllowOrigins []string // parsed allow-list actually applied by the CORS middleware
	JWTSecret        string
	SupabaseURL      string
	SupabaseKey      string
	APIFootballKey   string // API-Football (v3) key, used by the sync-players command
	DB               DBConfig
}

// DBConfig holds PostgreSQL connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	DSN      string // Full connection string override (DATABASE_URL)
}

// ConnString returns the PostgreSQL connection string.
// If DATABASE_URL is set, it takes precedence over individual fields.
func (c DBConfig) ConnString() string {
	if c.DSN != "" {
		return c.DSN
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Load reads configuration from environment variables with sensible defaults.
// It automatically loads .env if present, and returns an error if any required
// setting is missing — booting with an empty JWT_SECRET would let the server
// sign tokens with "" and make forgery trivial, so we fail fast instead.
func Load() (*Config, error) {
	_ = godotenv.Load() // silent fail if .env not found (e.g. in production)

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "debug"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
		// Prefer the plural CORS_ALLOW_ORIGINS; fall back to the legacy singular
		// CORS_ALLOW_ORIGIN; if neither is set, use the built-in default list.
		CORSAllowOrigins: parseCSV(getEnv("CORS_ALLOW_ORIGINS", getEnv("CORS_ALLOW_ORIGIN", defaultCORSOrigins))),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		SupabaseURL:      getEnv("SUPABASE_URL", ""),
		SupabaseKey:      getEnv("SUPABASE_SERVICE_KEY", ""),
		// Optional: only the sync-players command needs it, so we do not
		// fail-fast here when it is unset (unlike JWT_SECRET below).
		APIFootballKey: getEnv("API_FOOTBALL_KEY", ""),

		DB: DBConfig{
			DSN:      os.Getenv("DATABASE_URL"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "hypermotionleague"),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
		},
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set and non-empty")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseCSV splits a comma-separated string into a trimmed, non-empty slice.
func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
