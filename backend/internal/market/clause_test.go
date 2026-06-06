package market

import "testing"

func TestNewReleaseClause(t *testing.T) {
	tests := []struct {
		marketValue int
		want        int
	}{
		{1_000_000, 2_000_000},
		{0, 0},
		{-5, 0}, // guarded: never negative money
	}
	for _, tt := range tests {
		if got := NewReleaseClause(tt.marketValue); got != tt.want {
			t.Errorf("NewReleaseClause(%d) = %d, want %d", tt.marketValue, got, tt.want)
		}
	}
}

func TestReleaseClause(t *testing.T) {
	tests := []struct {
		name        string
		stored      int
		marketValue int
		want        int
	}{
		{"stored set wins", 5_000_000, 1_000_000, 5_000_000},
		{"stored 0 falls back to value*2", 0, 1_000_000, 2_000_000},
		{"negative stored falls back", -3, 100, 200},
	}
	for _, tt := range tests {
		if got := ReleaseClause(tt.stored, tt.marketValue); got != tt.want {
			t.Errorf("%s: ReleaseClause(%d, %d) = %d, want %d",
				tt.name, tt.stored, tt.marketValue, got, tt.want)
		}
	}
}

func TestApplyClausePaymentConservesMoney(t *testing.T) {
	cases := []struct{ buyer, seller, amount int }{
		{100, 50, 30},
		{1_000, 0, 1_000},
		{2_000_000, 5_000_000, 1_500_000},
		{60_000_000, 100_000_000, 30_000_000},
	}
	for _, c := range cases {
		nb, ns := ApplyClausePayment(c.buyer, c.seller, c.amount)
		if nb != c.buyer-c.amount {
			t.Errorf("buyer: got %d, want %d", nb, c.buyer-c.amount)
		}
		if ns != c.seller+c.amount {
			t.Errorf("seller: got %d, want %d", ns, c.seller+c.amount)
		}
		// The invariant that matters for money: the total is unchanged — the
		// clause is moved, never created or destroyed.
		if nb+ns != c.buyer+c.seller {
			t.Errorf("money not conserved: before %d, after %d",
				c.buyer+c.seller, nb+ns)
		}
	}
}
