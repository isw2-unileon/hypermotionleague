package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/market"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Release-clause operation errors. The handler maps each to an HTTP response.
var (
	// ErrClausePlayerFree: the player has no owner in the league, so there is no
	// clause to pay (sign them through the normal market instead).
	ErrClausePlayerFree = errors.New("player is a free agent: no clause to pay")
	// ErrClauseSelfPurchase: the buyer already owns the player. Also the guard
	// that makes a double-submit safe (see PayClauseTx).
	ErrClauseSelfPurchase = errors.New("buyer already owns this player")
	// ErrClauseInsufficient: the buyer's budget is below the clause.
	ErrClauseInsufficient = errors.New("insufficient budget for the release clause")
	// ErrClauseBuyerNotMember: the buyer is not a member of the league.
	ErrClauseBuyerNotMember = errors.New("buyer is not a member of this league")
)

// TeamRepo implements repository.TeamRepository for PostgreSQL.
type TeamRepo struct {
	pool *pgxpool.Pool
}

// NewTeamRepo creates a new TeamRepo.
func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

// AddPlayer adds a player to a user's team in a league.
func (r *TeamRepo) AddPlayer(ctx context.Context, tp *models.TeamPlayer) error {
	query := `
		INSERT INTO team_players (league_id, user_id, player_id, purchase_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, acquired_at`

	return r.pool.QueryRow(ctx, query,
		tp.LeagueID, tp.UserID, tp.PlayerID, tp.PurchasePrice,
	).Scan(&tp.ID, &tp.AcquiredAt)
}

// RemovePlayer removes a player from a user's team.
func (r *TeamRepo) RemovePlayer(ctx context.Context, leagueID, userID, playerID int64) error {
	query := `DELETE FROM team_players WHERE league_id = $1 AND user_id = $2 AND player_id = $3`
	_, err := r.pool.Exec(ctx, query, leagueID, userID, playerID)
	return err
}

// GetUserTeam retrieves a user's full squad in a league.
func (r *TeamRepo) GetUserTeam(ctx context.Context, leagueID, userID int64) (*models.UserTeam, error) {
	// Get user info and budget
	userQuery := `
		SELECT u.id, u.username, u.display_name, lm.budget
		FROM users u
		INNER JOIN league_members lm ON u.id = lm.user_id
		WHERE lm.league_id = $1 AND lm.user_id = $2`

	team := &models.UserTeam{LeagueID: leagueID, UserID: userID}
	err := r.pool.QueryRow(ctx, userQuery, leagueID, userID).Scan(
		&team.UserID, &team.Username, &team.DisplayName, &team.Budget,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user team info: %w", err)
	}

	// Get players
	playerQuery := `
		SELECT tp.id, tp.league_id, tp.user_id, tp.player_id, tp.purchase_price, tp.acquired_at,
		       p.id, p.first_name, p.last_name, p.position, p.team_name, p.market_value,
		       p.is_active, p.created_at, p.updated_at
		FROM team_players tp
		INNER JOIN players p ON tp.player_id = p.id
		WHERE tp.league_id = $1 AND tp.user_id = $2
		ORDER BY p.position, p.last_name`

	rows, err := r.pool.Query(ctx, playerQuery, leagueID, userID)
	if err != nil {
		return nil, fmt.Errorf("get team players: %w", err)
	}
	defer rows.Close()

	totalValue := 0
	for rows.Next() {
		var tpd models.TeamPlayerWithDetails
		err := rows.Scan(
			&tpd.ID, &tpd.LeagueID, &tpd.UserID, &tpd.PlayerID,
			&tpd.PurchasePrice, &tpd.AcquiredAt,
			&tpd.Player.ID, &tpd.Player.FirstName, &tpd.Player.LastName,
			&tpd.Player.Position, &tpd.Player.TeamName, &tpd.Player.MarketValue,
			&tpd.Player.IsActive, &tpd.Player.CreatedAt, &tpd.Player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan team player: %w", err)
		}
		team.Players = append(team.Players, tpd)
		totalValue += tpd.PurchasePrice
	}
	team.TotalValue = totalValue

	return team, rows.Err()
}

// GetPlayerOwner retrieves who owns a player in a league.
func (r *TeamRepo) GetPlayerOwner(ctx context.Context, leagueID, playerID int64) (*models.TeamPlayer, error) {
	query := `
		SELECT id, league_id, user_id, player_id, purchase_price, acquired_at
		FROM team_players
		WHERE league_id = $1 AND player_id = $2`

	tp := &models.TeamPlayer{}
	err := r.pool.QueryRow(ctx, query, leagueID, playerID).Scan(
		&tp.ID, &tp.LeagueID, &tp.UserID, &tp.PlayerID,
		&tp.PurchasePrice, &tp.AcquiredAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get player owner: %w", err)
	}
	return tp, nil
}

// HasPlayer checks if a user owns a specific player in a league.
func (r *TeamRepo) HasPlayer(ctx context.Context, leagueID, userID, playerID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM team_players WHERE league_id = $1 AND user_id = $2 AND player_id = $3)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, leagueID, userID, playerID).Scan(&exists)
	return exists, err
}

// TransferPlayer transfers a player between users in a league.
func (r *TeamRepo) TransferPlayer(ctx context.Context, leagueID, oldUserID, newUserID, playerID int64, price int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Remove from old user
	if _, err := tx.Exec(ctx,
		`DELETE FROM team_players WHERE league_id = $1 AND user_id = $2 AND player_id = $3`,
		leagueID, oldUserID, playerID,
	); err != nil {
		return fmt.Errorf("remove from old user: %w", err)
	}

	// Add to new user
	if _, err := tx.Exec(ctx,
		`INSERT INTO team_players (league_id, user_id, player_id, purchase_price) VALUES ($1, $2, $3, $4)`,
		leagueID, newUserID, playerID, price,
	); err != nil {
		return fmt.Errorf("add to new user: %w", err)
	}

	// Update budgets
	if _, err := tx.Exec(ctx,
		`UPDATE league_members SET budget = budget + $1 WHERE league_id = $2 AND user_id = $3`,
		price, leagueID, oldUserID,
	); err != nil {
		return fmt.Errorf("update seller budget: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE league_members SET budget = budget - $1 WHERE league_id = $2 AND user_id = $3`,
		price, leagueID, newUserID,
	); err != nil {
		return fmt.Errorf("update buyer budget: %w", err)
	}

	return tx.Commit(ctx)
}

// PayClauseTx executes a release-clause purchase: the buyer pays a player's
// clause and the player moves from its current owner to the buyer, with the
// clause amount transferred from the buyer's budget to the seller's — all in one
// transaction with FOR UPDATE locks, mirroring ResolveListingTx.
//
// The seller is whoever currently owns the player (discovered under the lock,
// not passed in by the caller), which is what makes a double-submit safe: the
// first call moves ownership to the buyer; a second concurrent/retried call,
// after locking the same team_players row, sees the buyer as the owner and
// returns ErrClauseSelfPurchase instead of charging again.
//
// Lock order matches the resolver: the team_players row first (FOR UPDATE OF
// tp), then the two league_members rows in ascending user_id order, which is
// deadlock-safe among concurrent clause purchases.
//
// The clause charged is the stored release_clause, or market_value × 2 as a
// fallback for legacy rows (see market.ReleaseClause). The player's new clause
// is recomputed from the current market_value so it does not keep the previous
// owner's clause.
func (r *TeamRepo) PayClauseTx(ctx context.Context, leagueID, buyerID, playerID int64) (*models.ClauseResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Lock the ownership row and read the current owner + pricing inputs.
	var (
		tpID         int64
		sellerID     int64
		storedClause *int64 // NULL for legacy rows
		marketValue  int
		firstName    string
		lastName     string
	)
	err = tx.QueryRow(ctx, `
		SELECT tp.id, tp.user_id, tp.release_clause, p.market_value, p.first_name, p.last_name
		FROM team_players tp
		INNER JOIN players p ON p.id = tp.player_id
		WHERE tp.league_id = $1 AND tp.player_id = $2
		FOR UPDATE OF tp
	`, leagueID, playerID).Scan(&tpID, &sellerID, &storedClause, &marketValue, &firstName, &lastName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClausePlayerFree
		}
		return nil, fmt.Errorf("lock ownership row: %w", err)
	}

	// 2. A buyer cannot pay the clause of a player they already own. This also
	//    catches a double-submit: after the first purchase the buyer is owner.
	if sellerID == buyerID {
		return nil, ErrClauseSelfPurchase
	}

	// 3. Lock both member rows in ascending user_id order (deadlock-safe).
	budgets := make(map[int64]int, 2)
	rows, err := tx.Query(ctx, `
		SELECT user_id, budget FROM league_members
		WHERE league_id = $1 AND user_id IN ($2, $3)
		ORDER BY user_id
		FOR UPDATE
	`, leagueID, buyerID, sellerID)
	if err != nil {
		return nil, fmt.Errorf("lock member rows: %w", err)
	}
	for rows.Next() {
		var uid int64
		var b int
		if err := rows.Scan(&uid, &b); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan member budget: %w", err)
		}
		budgets[uid] = b
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate member rows: %w", err)
	}

	buyerBudget, ok := budgets[buyerID]
	if !ok {
		return nil, ErrClauseBuyerNotMember
	}
	sellerBudget, ok := budgets[sellerID]
	if !ok {
		// The owner should always be a member; treat a missing row as data
		// inconsistency rather than silently proceeding.
		return nil, fmt.Errorf("seller %d is not a member of league %d", sellerID, leagueID)
	}

	// 4. Compute the clause (stored, or market_value × 2 fallback) and validate.
	stored := 0
	if storedClause != nil {
		stored = int(*storedClause)
	}
	clause := market.ReleaseClause(stored, marketValue)
	if buyerBudget < clause {
		return nil, ErrClauseInsufficient
	}
	newClause := market.NewReleaseClause(marketValue)
	newBuyerBudget, newSellerBudget := market.ApplyClausePayment(buyerBudget, sellerBudget, clause)

	// 5. Reassign ownership to the buyer, set the price paid and the new clause.
	//    UPDATE (not delete+insert) keeps uq_team_player satisfied throughout.
	if _, err = tx.Exec(ctx, `
		UPDATE team_players
		SET user_id = $1, purchase_price = $2, release_clause = $3, acquired_at = NOW()
		WHERE id = $4
	`, buyerID, clause, newClause, tpID); err != nil {
		return nil, fmt.Errorf("reassign player: %w", err)
	}

	// 6. Move the money: charge the buyer, credit the seller (relative updates,
	//    same style as ResolveListingTx; rows are already locked above).
	if _, err = tx.Exec(ctx, `
		UPDATE league_members SET budget = budget - $1 WHERE league_id = $2 AND user_id = $3
	`, clause, leagueID, buyerID); err != nil {
		return nil, fmt.Errorf("charge buyer: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		UPDATE league_members SET budget = budget + $1 WHERE league_id = $2 AND user_id = $3
	`, clause, leagueID, sellerID); err != nil {
		return nil, fmt.Errorf("credit seller: %w", err)
	}

	// 7. Remove the player from the seller's not-yet-finished lineups, so he is
	//    not fielded by his former owner. Past (already-played) matchdays are
	//    left untouched to preserve scored history.
	if _, err = tx.Exec(ctx, `
		DELETE FROM lineup_players
		WHERE player_id = $1
		  AND lineup_id IN (
		      SELECT l.id FROM lineups l
		      INNER JOIN matchdays m ON m.id = l.matchday_id
		      WHERE l.league_id = $2 AND l.user_id = $3 AND m.end_date > NOW()
		  )
	`, playerID, leagueID, sellerID); err != nil {
		return nil, fmt.Errorf("remove from seller lineups: %w", err)
	}

	result := &models.ClauseResult{
		LeagueID:         leagueID,
		PlayerID:         playerID,
		PlayerName:       firstName + " " + lastName,
		PreviousOwnerID:  sellerID,
		NewOwnerID:       buyerID,
		AmountPaid:       clause,
		NewReleaseClause: newClause,
		BuyerBudget:      newBuyerBudget,
		SellerBudget:     newSellerBudget,
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit clause payment: %w", err)
	}
	return result, nil
}

// draftQuota describes how many players of a position make up an initial squad.
type draftQuota struct {
	position models.PlayerPosition
	count    int
}

// initialSquadQuotas defines the fixed shape of an initial draft squad:
// 2 GK, 5 DEF, 5 MID, 3 FWD "15 players"
var initialSquadQuotas = []draftQuota{
	{models.PositionGK, 2},
	{models.PositionDEF, 5},
	{models.PositionMID, 5},
	{models.PositionFWD, 3},
}

// DraftInitialSquad assigns a random starting squad of 15 players (2 GK, 5 DEF,
// 5 MID, 3 FWD) to a user in a league.
func (r *TeamRepo) DraftInitialSquad(ctx context.Context, leagueID, userID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, q := range initialSquadQuotas {
		rows, err := tx.Query(ctx,
			`SELECT id FROM players
			 WHERE is_active = TRUE
			   AND position = $1
			   AND id NOT IN (SELECT player_id FROM team_players WHERE league_id = $2)
			 ORDER BY RANDOM()
			 LIMIT $3`,
			q.position, leagueID, q.count,
		)
		if err != nil {
			return fmt.Errorf("select available %s: %w", q.position, err)
		}

		playerIDs := make([]int64, 0, q.count)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return fmt.Errorf("scan available %s: %w", q.position, err)
			}
			playerIDs = append(playerIDs, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate available %s: %w", q.position, err)
		}

		if len(playerIDs) < q.count {
			return fmt.Errorf("not enough available %s in this league: need %d, found %d",
				q.position, q.count, len(playerIDs))
		}

		for _, playerID := range playerIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO team_players (league_id, user_id, player_id, purchase_price)
				 VALUES ($1, $2, $3, 0)`,
				leagueID, userID, playerID,
			); err != nil {
				return fmt.Errorf("assign %s player %d: %w", q.position, playerID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit draft: %w", err)
	}
	return nil
}
