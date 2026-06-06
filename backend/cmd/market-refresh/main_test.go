package main

import (
	"math/rand"
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

func mkPlayer(id int64, pos models.PlayerPosition, value int) models.Player {
	return models.Player{ID: id, Position: pos, MarketValue: value}
}

// pool builds a free-agent set with the given counts per position; market_value
// decreases with id so values are distinct.
func pool(gk, def, mid, fwd int) []models.Player {
	var ps []models.Player
	var id int64
	add := func(pos models.PlayerPosition, n int) {
		for i := 0; i < n; i++ {
			id++
			ps = append(ps, mkPlayer(id, pos, 1_000_000+int(1000-id)*1000))
		}
	}
	add(models.PositionGK, gk)
	add(models.PositionDEF, def)
	add(models.PositionMID, mid)
	add(models.PositionFWD, fwd)
	return ps
}

func countByPos(players []models.Player) map[models.PlayerPosition]int {
	m := map[models.PlayerPosition]int{}
	for _, p := range players {
		m[p.Position]++
	}
	return m
}

func TestAllocateQuotas(t *testing.T) {
	tests := []struct {
		n    int
		want map[models.PlayerPosition]int
	}{
		{12, map[models.PlayerPosition]int{models.PositionGK: 2, models.PositionDEF: 4, models.PositionMID: 4, models.PositionFWD: 2}},
		{6, map[models.PlayerPosition]int{models.PositionGK: 1, models.PositionDEF: 2, models.PositionMID: 2, models.PositionFWD: 1}},
	}
	for _, tt := range tests {
		got := allocateQuotas(tt.n)
		sum := 0
		for _, v := range got {
			sum += v
		}
		if sum != tt.n {
			t.Errorf("allocateQuotas(%d) sums to %d, want %d (%v)", tt.n, sum, tt.n, got)
		}
		for pos, want := range tt.want {
			if got[pos] != want {
				t.Errorf("allocateQuotas(%d)[%s] = %d, want %d", tt.n, pos, got[pos], want)
			}
		}
	}

	// For any n, the quotas must always sum to exactly n.
	for n := 1; n <= 20; n++ {
		got := allocateQuotas(n)
		sum := 0
		for _, v := range got {
			sum += v
		}
		if sum != n {
			t.Errorf("allocateQuotas(%d) sums to %d, want %d", n, sum, n)
		}
	}
}

func TestSelectFreeAgents_Distribution(t *testing.T) {
	free := pool(5, 10, 10, 5) // plenty of every position
	rng := rand.New(rand.NewSource(1))

	got := selectFreeAgents(free, 12, rng)
	if len(got) != 12 {
		t.Fatalf("len = %d, want 12", len(got))
	}
	assertUnique(t, got)

	by := countByPos(got)
	want := map[models.PlayerPosition]int{models.PositionGK: 2, models.PositionDEF: 4, models.PositionMID: 4, models.PositionFWD: 2}
	for pos, w := range want {
		if by[pos] != w {
			t.Errorf("position %s = %d, want %d", pos, by[pos], w)
		}
	}
}

func TestSelectFreeAgents_ShortPositionFills(t *testing.T) {
	// Only 1 GK available: GK quota is 2, so the shortfall must be filled from
	// other positions and the total must still reach 12.
	free := pool(1, 12, 12, 6)
	rng := rand.New(rand.NewSource(7))

	got := selectFreeAgents(free, 12, rng)
	if len(got) != 12 {
		t.Fatalf("len = %d, want 12", len(got))
	}
	assertUnique(t, got)
	if c := countByPos(got)[models.PositionGK]; c != 1 {
		t.Errorf("GK count = %d, want 1 (only one available)", c)
	}
}

func TestSelectFreeAgents_FewerThanTarget(t *testing.T) {
	free := pool(1, 1, 1, 1) // only 4 free agents, target 12
	rng := rand.New(rand.NewSource(3))

	got := selectFreeAgents(free, 12, rng)
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4 (all available)", len(got))
	}
	assertUnique(t, got)
}

func TestSelectFreeAgents_Deterministic(t *testing.T) {
	free := pool(5, 10, 10, 5)
	a := selectFreeAgents(free, 12, rand.New(rand.NewSource(42)))
	b := selectFreeAgents(free, 12, rand.New(rand.NewSource(42)))
	if !sameIDs(a, b) {
		t.Errorf("same seed produced different selections")
	}
}

func TestSelectFreeAgents_Rotates(t *testing.T) {
	// Over many draws against a large pool, the market must show more than 12
	// distinct players — i.e. it rotates rather than always offering the same.
	free := pool(8, 16, 16, 8)
	rng := rand.New(rand.NewSource(99))

	seen := map[int64]bool{}
	for i := 0; i < 15; i++ {
		for _, p := range selectFreeAgents(free, 12, rng) {
			seen[p.ID] = true
		}
	}
	if len(seen) <= 12 {
		t.Errorf("only %d distinct players ever offered across 15 draws; expected rotation (>12)", len(seen))
	}
}

func assertUnique(t *testing.T, players []models.Player) {
	t.Helper()
	seen := map[int64]bool{}
	for _, p := range players {
		if seen[p.ID] {
			t.Errorf("duplicate player id %d in selection", p.ID)
		}
		seen[p.ID] = true
	}
}

func sameIDs(a, b []models.Player) bool {
	if len(a) != len(b) {
		return false
	}
	sa := map[int64]bool{}
	for _, p := range a {
		sa[p.ID] = true
	}
	for _, p := range b {
		if !sa[p.ID] {
			return false
		}
	}
	return true
}
