// Package main is the entry point for the backend server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/handlers"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/middleware"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	ctx := context.Background()

	cfg := config.Load()

	// Initialize database connection
	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Initialize repositories
	repos := postgres.NewRepositories(pool.Pool)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(repos.User, cfg.JWTSecret)
	oauthHandler := handlers.NewOAuthHandler(repos.User, cfg.JWTSecret, cfg.SupabaseURL, cfg.SupabaseKey)
	leagueHandler := handlers.NewLeagueHandler(repos.League)
	matchdayHandler := handlers.NewMatchdayHandler(repos.Matchday)
	playerHandler := handlers.NewPlayerHandler(repos.Player, repos.Matchday)
	teamHandler := handlers.NewTeamHandler(repos.Team, repos.League)
	lineupHandler := handlers.NewLineupHandler(repos.Matchday, repos.Team, repos.League)
	marketHandler := handlers.NewMarketHandler(repos.Market, repos.Player, repos.Team, repos.League)
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Public routes, no auth required
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
	})

	// GET /api/db-test checks if the database connection is alive
	api.GET("/db-test", func(c *gin.Context) {
		if err := pool.Pool.Ping(c.Request.Context()); err != nil {
			logger.Error("database ping failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "DB conectada"})
	})

	// Auth — public routes (register and login do not require token)
	v1 := api.Group("/v1")
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/oauth", oauthHandler.Callback)

	// Players: this is public no auth required
	v1.GET("/players/:id", playerHandler.GetByID)
	v1.GET("players", playerHandler.List)
	v1.GET("/players/:id/points", playerHandler.GetPointsByMatchday)

	// Protected routes — from here on, all require a valid JWT
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		protected.GET("/leagues", leagueHandler.GetByUserID)
		protected.POST("/leagues", leagueHandler.Create)
		protected.POST("/leagues/join", leagueHandler.JoinLeague)
		protected.GET("/leagues/:id", leagueHandler.GetByID)
		protected.GET("/leagues/:id/members", leagueHandler.GetMembers)
		protected.DELETE("/leagues/:id", leagueHandler.Delete)

		protected.GET("/leagues/:id/matchdays", matchdayHandler.GetByLeague)
		protected.GET("/leagues/:id/matchdays/current", matchdayHandler.GetCurrent)
		protected.GET("/leagues/:id/standings", matchdayHandler.GetStandings)
		protected.GET("/leagues/:id/matchdays/:number/standings", matchdayHandler.GetMatchdayStandings)

		// Team
		protected.GET("/leagues/:id/team", teamHandler.GetUserTeam)

		// Lineup
		protected.GET("/leagues/:id/matchdays/:number/lineup", lineupHandler.GetLineup)
		protected.PUT("/leagues/:id/matchdays/:number/lineup", lineupHandler.SaveLineup)
		protected.DELETE("/leagues/:id/matchdays/:number/lineup/players/:player_id", lineupHandler.RemoveLineupPlayer)
		// Market
		protected.GET("/leagues/:id/market/players", marketHandler.GetAvailablePlayers)
		protected.GET("/leagues/:id/market/listings", marketHandler.GetActiveListings)
		protected.POST("/leagues/:id/market/bids", marketHandler.PlaceBid)
		protected.GET("/leagues/:id/market/bids", marketHandler.GetUserBids)
		protected.DELETE("/leagues/:id/market/bids/:bid_id", marketHandler.CancelBid)
		protected.GET("/leagues/:id/market/status", marketHandler.GetMarketStatus)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("server stopped")
}
