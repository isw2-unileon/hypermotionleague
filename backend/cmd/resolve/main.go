// Package main provides a one-shot CLI that resolves all expired market
// listings: awards the player to the highest bidder (deducting budget),
// marks losing bids, and closes the listing. Listings with no bids are
// simply closed.
//
// Usage:
//
//	go run ./backend/cmd/resolve
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	marketRepo := postgres.NewMarketRepo(pool.Pool)

	// 1. Fetch all expired active listings
	listings, err := marketRepo.GetExpiredListings(ctx)
	if err != nil {
		logger.Error("failed to get expired listings", "error", err)
		os.Exit(1)
	}

	if len(listings) == 0 {
		logger.Info("no expired listings to resolve")
		return
	}

	logger.Info("found expired listings", "count", len(listings))

	// 2. Resolve each listing in its own transaction
	resolved := 0
	failed := 0
	for _, listing := range listings {
		if err := marketRepo.ResolveListingTx(ctx, listing); err != nil {
			logger.Error("failed to resolve listing",
				"listing_id", listing.ID,
				"player_id", listing.PlayerID,
				"league_id", listing.LeagueID,
				"error", err,
			)
			failed++
			continue
		}
		logger.Info("resolved listing",
			"listing_id", listing.ID,
			"player_id", listing.PlayerID,
			"league_id", listing.LeagueID,
		)
		resolved++
	}

	logger.Info("resolution complete",
		"resolved", resolved,
		"failed", failed,
		"total", len(listings),
	)
}
