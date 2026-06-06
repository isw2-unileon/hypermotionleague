// Command market-refresh repopulates each league's market with free-agent
// players, so the market is never empty. It is a one-shot job meant to be fired
// daily when the market opens (by a cron or by hand) — it intentionally does NOT
// embed a scheduler and does NOT know about trading hours; deciding WHEN to run
// (e.g. 19:00) belongs to whatever triggers it, not here. It mirrors the one-shot
// shape of cmd/resolve and cmd/sync-players.
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
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// listingTTL is how long a refreshed listing lives — one day, until the next
// refresh. The resolver (cmd/resolve) closes them once expired.
const listingTTL = 24 * time.Hour

// positionQuota is the target spread of a full (12-listing) market across
// positions. Quotas for a different target are derived proportionally (see
// allocateQuotas). Weights: 2 GK, 4 DEF, 4 MID, 2 FWD = 12.
var positionQuota = []struct {
	pos    models.PlayerPosition
	weight int
}{
	{models.PositionGK, 2},
	{models.PositionDEF, 4},
	{models.PositionMID, 4},
	{models.PositionFWD, 2},
}

func main() {
	dryRun := flag.Bool("dry-run", true,
		"compute and log what WOULD be created, writing nothing; pass --dry-run=false to persist")
	perLeague := flag.Int("per-league", 12, "target number of active listings per league")
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

	// Per-process RNG seeded from the clock drives the value-biased rotation, so
	// each daily run offers a different mix (see selectFreeAgents).
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	expiresAt := time.Now().Add(listingTTL)

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

// refreshLeague tops one league up to perLeague active listings. It returns how
// many listings were created (or would be, in dry-run) and whether the league
// was short of free agents to reach the target.
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
	active, err := marketRepo.CountActiveListings(ctx, leagueID)
	if err != nil {
		return 0, false, err
	}

	needed := perLeague - active
	if needed <= 0 {
		logger.Info("league already stocked",
			"league_id", leagueID, "active", active, "target", perLeague)
		return 0, false, nil
	}

	free, err := playerRepo.GetFreeAgentsForLeague(ctx, leagueID)
	if err != nil {
		return 0, false, err
	}

	chosen := selectFreeAgents(free, needed, rng)
	short = len(chosen) < needed // not enough free agents to reach the target

	listings := make([]models.MarketListing, 0, len(chosen))
	for _, p := range chosen {
		listings = append(listings, models.MarketListing{
			LeagueID:  leagueID,
			PlayerID:  p.ID,
			BasePrice: p.MarketValue, // base_price = player's market_value
			SellerID:  nil,           // system offer, no seller
			Status:    "active",
			ExpiresAt: expiresAt,
		})
	}

	breakdown := positionBreakdown(chosen)

	if dryRun {
		logger.Info("would create listings",
			"league_id", leagueID, "active_now", active, "target", perLeague,
			"to_create", len(listings), "free_agents", len(free),
			"by_position", breakdown, "short", short)
		for _, p := range chosen {
			logger.Info("  would list",
				"league_id", leagueID, "player_id", p.ID, "name", p.FullName(),
				"position", p.Position, "price", p.MarketValue)
		}
		return len(listings), short, nil
	}

	inserted, err := marketRepo.InsertListingsTx(ctx, listings)
	if err != nil {
		return 0, short, err
	}
	logger.Info("created listings",
		"league_id", leagueID, "active_now", active, "target", perLeague,
		"created", inserted, "free_agents", len(free),
		"by_position", breakdown, "short", short)
	return inserted, short, nil
}

// selectFreeAgents picks up to n players from the free agents, spread across
// positions (see positionQuota) with value-biased rotation. It is pure: the rng
// is injected so the choice is testable and so each run differs.
//
// Within a position, players are sampled WITHOUT replacement weighted by
// market_value (Efraimidis–Spirakis: key = u^(1/value), pick the largest keys),
// so higher-valued players are more likely but not guaranteed — the market
// rotates day to day instead of always showing the same faces. If a position has
// fewer free agents than its quota, the shortfall is filled from the remaining
// players of any position (also value-weighted).
func selectFreeAgents(free []models.Player, n int, rng *rand.Rand) []models.Player {
	if n <= 0 || len(free) == 0 {
		return nil
	}
	if n >= len(free) {
		out := make([]models.Player, len(free))
		copy(out, free)
		return out
	}

	groups := make(map[models.PlayerPosition][]models.Player)
	for _, p := range free {
		groups[p.Position] = append(groups[p.Position], p)
	}

	quotas := allocateQuotas(n)
	chosen := make([]models.Player, 0, n)
	picked := make(map[int64]bool, n)
	for _, pq := range positionQuota {
		for _, p := range weightedSample(groups[pq.pos], quotas[pq.pos], rng) {
			chosen = append(chosen, p)
			picked[p.ID] = true
		}
	}

	// Fill the shortfall (positions that lacked enough free agents) from the
	// remaining players of any position.
	if len(chosen) < n {
		var rest []models.Player
		for _, p := range free {
			if !picked[p.ID] {
				rest = append(rest, p)
			}
		}
		chosen = append(chosen, weightedSample(rest, n-len(chosen), rng)...)
	}

	return chosen
}

// allocateQuotas splits n slots across the positions in positionQuota,
// proportional to their weights, using largest-remainder apportionment so the
// quotas always sum to exactly n. For n == 12 it yields {GK:2, DEF:4, MID:4, FWD:2}.
func allocateQuotas(n int) map[models.PlayerPosition]int {
	total := 0
	for _, pq := range positionQuota {
		total += pq.weight
	}

	type rem struct {
		pos  models.PlayerPosition
		frac float64
	}
	quotas := make(map[models.PlayerPosition]int, len(positionQuota))
	rems := make([]rem, 0, len(positionQuota))
	assigned := 0
	for _, pq := range positionQuota {
		exact := float64(n) * float64(pq.weight) / float64(total)
		base := int(math.Floor(exact))
		quotas[pq.pos] = base
		assigned += base
		rems = append(rems, rem{pq.pos, exact - float64(base)})
	}

	// Hand out the leftover slots to the largest fractional remainders.
	sort.Slice(rems, func(i, j int) bool { return rems[i].frac > rems[j].frac })
	for i := 0; i < n-assigned && i < len(rems); i++ {
		quotas[rems[i].pos]++
	}
	return quotas
}

// weightedSample returns up to k players sampled without replacement, weighted by
// market_value (higher value -> more likely). Uses the Efraimidis–Spirakis key
// u^(1/weight) and keeps the largest keys.
func weightedSample(players []models.Player, k int, rng *rand.Rand) []models.Player {
	if k <= 0 || len(players) == 0 {
		return nil
	}
	if k >= len(players) {
		out := make([]models.Player, len(players))
		copy(out, players)
		return out
	}

	type keyed struct {
		p   models.Player
		key float64
	}
	ks := make([]keyed, len(players))
	for i, p := range players {
		w := float64(p.MarketValue)
		if w < 1 {
			w = 1 // guard zero/negative so the key is well-defined
		}
		u := 1.0 - rng.Float64() // (0,1], avoids log/pow of 0
		ks[i] = keyed{p, math.Pow(u, 1.0/w)}
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i].key > ks[j].key })

	out := make([]models.Player, k)
	for i := 0; i < k; i++ {
		out[i] = ks[i].p
	}
	return out
}

// positionBreakdown counts the chosen players per position, for logging.
func positionBreakdown(players []models.Player) map[string]int {
	out := make(map[string]int, len(positionQuota))
	for _, p := range players {
		out[string(p.Position)]++
	}
	return out
}
