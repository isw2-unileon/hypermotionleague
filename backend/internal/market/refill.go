package market

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// DefaultTarget is the number of active listings a league's market is topped up
// to. Shared by the daily refresh command and by on-create league seeding so
// both aim at the same stocked market.
const DefaultTarget = 12

// ListingTTL is how long a refreshed listing lives — one day, until the next
// refresh. The resolver (cmd/resolve) closes them once expired.
const ListingTTL = 24 * time.Hour

// ListingStore is the market-side persistence the refill needs: how many active
// listings a league has, and how to insert new ones idempotently. Satisfied by
// *postgres.MarketRepo. Declared here (rather than importing postgres) so this
// package stays free of an import cycle (postgres already imports market).
type ListingStore interface {
	CountActiveListings(ctx context.Context, leagueID int64) (int, error)
	InsertListingsTx(ctx context.Context, listings []models.MarketListing) (int, error)
}

// FreeAgentSource yields the players eligible to be listed in a league.
// Satisfied by *postgres.PlayerRepo.
type FreeAgentSource interface {
	GetFreeAgentsForLeague(ctx context.Context, leagueID int64) ([]models.Player, error)
}

// positionQuota is the target spread of a full (DefaultTarget) market across
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

// RefillPlan is the computed (read-only) plan to top a league's market up to a
// target: which free agents were chosen and the listings to insert for them.
type RefillPlan struct {
	Target       int
	ActiveBefore int                    // active listings already in the league
	FreeAgents   int                    // free agents considered
	Chosen       []models.Player        // the selected players
	Listings     []models.MarketListing // one listing per chosen player, ready to insert
	Short        bool                   // true if there were not enough free agents to reach Target
}

// PlanRefill computes how to top a league up to `target` active listings WITHOUT
// writing anything: it counts current active listings, fetches the free agents,
// and selects up to the shortfall spread across positions with value-biased
// rotation (see SelectFreeAgents). Persisting is the caller's job (via
// ListingStore.InsertListingsTx), so the same plan can be previewed (dry-run) or
// applied. If the league is already at/above target, Listings is empty.
func PlanRefill(ctx context.Context, store ListingStore, src FreeAgentSource, leagueID int64, target int, expiresAt time.Time, rng *rand.Rand) (RefillPlan, error) {
	active, err := store.CountActiveListings(ctx, leagueID)
	if err != nil {
		return RefillPlan{}, fmt.Errorf("count active listings: %w", err)
	}

	plan := RefillPlan{Target: target, ActiveBefore: active}
	needed := target - active
	if needed <= 0 {
		return plan, nil
	}

	free, err := src.GetFreeAgentsForLeague(ctx, leagueID)
	if err != nil {
		return RefillPlan{}, fmt.Errorf("get free agents: %w", err)
	}
	plan.FreeAgents = len(free)

	chosen := SelectFreeAgents(free, needed, rng)
	plan.Chosen = chosen
	plan.Short = len(chosen) < needed

	plan.Listings = make([]models.MarketListing, 0, len(chosen))
	for _, p := range chosen {
		plan.Listings = append(plan.Listings, models.MarketListing{
			LeagueID:  leagueID,
			PlayerID:  p.ID,
			BasePrice: p.MarketValue, // base_price = player's market_value
			SellerID:  nil,           // system offer, no seller
			Status:    "active",
			ExpiresAt: expiresAt,
		})
	}
	return plan, nil
}

// RefreshLeague plans the refill and persists it in one call, returning how many
// listings were actually inserted (idempotent via InsertListingsTx) and the
// plan. Used by league creation; the daily command uses PlanRefill directly so
// it can also preview in dry-run.
func RefreshLeague(ctx context.Context, store ListingStore, src FreeAgentSource, leagueID int64, target int, expiresAt time.Time, rng *rand.Rand) (created int, plan RefillPlan, err error) {
	plan, err = PlanRefill(ctx, store, src, leagueID, target, expiresAt, rng)
	if err != nil {
		return 0, plan, err
	}
	if len(plan.Listings) == 0 {
		return 0, plan, nil
	}
	created, err = store.InsertListingsTx(ctx, plan.Listings)
	if err != nil {
		return 0, plan, fmt.Errorf("insert listings: %w", err)
	}
	return created, plan, nil
}

// SelectFreeAgents picks up to n players from the free agents, spread across
// positions (see positionQuota) with value-biased rotation. It is pure: the rng
// is injected so the choice is testable and so each run differs.
//
// Within a position, players are sampled WITHOUT replacement weighted by
// market_value (Efraimidis–Spirakis: key = u^(1/value), pick the largest keys),
// so higher-valued players are more likely but not guaranteed — the market
// rotates day to day instead of always showing the same faces. If a position has
// fewer free agents than its quota, the shortfall is filled from the remaining
// players of any position (also value-weighted).
func SelectFreeAgents(free []models.Player, n int, rng *rand.Rand) []models.Player {
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

// PositionBreakdown counts players per position, for logging.
func PositionBreakdown(players []models.Player) map[string]int {
	out := make(map[string]int, len(positionQuota))
	for _, p := range players {
		out[string(p.Position)]++
	}
	return out
}
