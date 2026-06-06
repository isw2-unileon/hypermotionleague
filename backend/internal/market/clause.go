package market

// Release-clause rules, kept here (next to the window rule) as the single source
// of truth so the SQL never re-encodes the ×2 factor. Pure functions: no DB, no
// clock — unit-testable, including the money-conservation invariant.

// ReleaseClauseFactor multiplies a player's market value to get their default
// release clause. Buying a rival's player costs double his market value, so it
// is not trivial to poach.
const ReleaseClauseFactor = 2

// NewReleaseClause returns the default clause for a player given his current
// market value (market_value × ReleaseClauseFactor). Never negative.
func NewReleaseClause(marketValue int) int {
	if marketValue <= 0 {
		return 0
	}
	return marketValue * ReleaseClauseFactor
}

// ReleaseClause returns the clause to charge for an owned player: the stored
// clause if it is set (> 0), otherwise the fallback market_value × factor for
// legacy rows that predate the release_clause column (stored 0 or NULL -> 0).
func ReleaseClause(stored, marketValue int) int {
	if stored > 0 {
		return stored
	}
	return NewReleaseClause(marketValue)
}

// ApplyClausePayment returns the buyer's and seller's budgets after the buyer
// pays `amount` to the seller. Pure helper so the money math — and its
// conservation (nothing is created or destroyed) — is unit-testable.
func ApplyClausePayment(buyerBudget, sellerBudget, amount int) (newBuyer, newSeller int) {
	return buyerBudget - amount, sellerBudget + amount
}
