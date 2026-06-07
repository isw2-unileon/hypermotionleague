// Command market-refresh repopulates each league's market with free-agent
// players, so the market is never empty. It is a one-shot job meant to be fired
// daily when the market opens (by a cron or by hand) — it intentionally does NOT
// embed a scheduler and does NOT know about trading hours; deciding WHEN to run
// (e.g. 19:00) belongs to whatever triggers it, not here. It mirrors the one-shot
// shape of cmd/resolve and cmd/sync-players.
//
// The per-league selection/insertion logic lives in the market package
// (market.PlanRefill / market.RefreshLeague) so it is shared with the on-create
// league market seeding; this command is just the CLI wrapper (flags, looping
// over leagues, dry-run, logging).
//
// For each league it tops up the active listings to a target (default 12),
// choosing free agents (not owned in the league, real players with an
// external_id, no existing active listing) spread across positions with
// value-biased rotation so the market changes day to day. base_price is the
// player's market_value, seller_id is NULL (system offer), expires_at is 24h out.
//
// It is dry-run by default: with no flags it logs what it WOULD create and writes
// nothing. Pass --dry-run=false to persist. Each league is processed in its own
// transaction, so a failure in one does not affect the others.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/market"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	dryRun := flag.Bool("dry-run", true,
		"compute and log what WOULD be created, writing nothing; pass --dry-run=false to persist")
	perLeague := flag.Int("per-league", market.DefaultTarget, "target number of active listings per league")
	league := flag.Int64("league", 0, "limit to a single league id (0 = all leagues)")
	flag.Parse()

	if *perLeague <= 0 {
		logger.Error("invalid flags", "error", "--per-league must be > 0")
		os.Exit(2)
	}

	if err := run(*dryRun, *perLeague, *league); err != nil {
		logger.Error("market-refresh failed", "error", err)
		os.Exit(1)
	}
}

func run(dryRun bool, perLeague int, onlyLeague int64) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	leagueRepo := postgres.NewLeagueRepo(pool.Pool)
	playerRepo := postgres.NewPlayerRepo(pool.Pool)
	marketRepo := postgres.NewMarketRepo(pool.Pool)

	// #nosec G404 -- el mercado no necesita aleatoriedad criptográfica;
	// math/rand es suficiente para barajar la rotación diaria de jugadores.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	expiresAt := time.Now().Add(market.ListingTTL)

	var leagueIDs []int64
	if onlyLeague > 0 {
		leagueIDs = []int64{onlyLeague}
	} else {
		leagueIDs, err = leagueRepo.ListIDs(ctx)
		if err != nil {
			return fmt.Errorf("list leagues: %w", err)
		}
	}

	logger.Info("starting market-refresh",
		"dry_run", dryRun, "per_league", perLeague, "leagues", len(leagueIDs))

	var processed, failed, totalCreated, shortLeagues int
	for _, lid := range leagueIDs {
		created, short, err := refreshLeague(ctx, marketRepo, playerRepo, lid, perLeague, expiresAt, rng, dryRun)
		if err != nil {
			// Per-league failure: log and continue so the other leagues still run.
			logger.Error("league refresh failed", "league_id", lid, "error", err)
			failed++
			continue
		}
		processed++
		totalCreated += created
		if short {
			shortLeagues++
		}
	}

	logger.Info("market-refresh complete",
		"dry_run", dryRun,
		"leagues_processed", processed,
		"leagues_failed", failed,
		"listings_created", totalCreated,
		"leagues_short", shortLeagues,
	)

	if failed > 0 {
		return fmt.Errorf("%d of %d leagues failed to refresh", failed, len(leagueIDs))
	}
	return nil
}

// refreshLeague tops one league up to perLeague active listings using the shared
// market.PlanRefill logic, then (unless dry-run) persists via InsertListingsTx.
// It returns how many listings were created (or would be, in dry-run) and
// whether the league was short of free agents to reach the target.
func refreshLeague(
	ctx context.Context,
	marketRepo *postgres.MarketRepo,
	playerRepo *postgres.PlayerRepo,
	leagueID int64,
	perLeague int,
	expiresAt time.Time,
	rng *rand.Rand,
	dryRun bool,
) (created int, short bool, err error) {
	plan, err := market.PlanRefill(ctx, marketRepo, playerRepo, leagueID, perLeague, expiresAt, rng)
	if err != nil {
		return 0, false, err
	}

	if len(plan.Listings) == 0 {
		logger.Info("league already stocked",
			"league_id", leagueID, "active", plan.ActiveBefore, "target", perLeague)
		return 0, false, nil
	}

	if dryRun {
		logger.Info("would create listings",
			"league_id", leagueID, "active_now", plan.ActiveBefore, "target", perLeague,
			"to_create", len(plan.Listings), "free_agents", plan.FreeAgents,
			"by_position", market.PositionBreakdown(plan.Chosen), "short", plan.Short)
		for _, p := range plan.Chosen {
			logger.Info("  would list",
				"league_id", leagueID, "player_id", p.ID, "name", p.FullName(),
				"position", p.Position, "price", p.MarketValue)
		}
		return len(plan.Listings), plan.Short, nil
	}

	inserted, err := marketRepo.InsertListingsTx(ctx, plan.Listings)
	if err != nil {
		return 0, plan.Short, err
	}
	logger.Info("created listings",
		"league_id", leagueID, "active_now", plan.ActiveBefore, "target", perLeague,
		"created", inserted, "free_agents", plan.FreeAgents,
		"by_position", market.PositionBreakdown(plan.Chosen), "short", plan.Short)
	return inserted, plan.Short, nil
}
