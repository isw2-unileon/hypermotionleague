package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	defer tx.Rollback(ctx)

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
