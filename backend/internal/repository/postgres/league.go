package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeagueRepo implements repository.LeagueRepository for PostgreSQL.
type LeagueRepo struct {
	pool *pgxpool.Pool
}

// NewLeagueRepo creates a new LeagueRepo.
func NewLeagueRepo(pool *pgxpool.Pool) *LeagueRepo {
	return &LeagueRepo{pool: pool}
}

// Create inserts a new league.
func (r *LeagueRepo) Create(ctx context.Context, league *models.League) error {
	query := `
		INSERT INTO leagues (name, invite_code, max_members, budget_per_user, market_close_time, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		league.Name, league.InviteCode, league.MaxMembers,
		league.BudgetPerUser, league.MarketCloseTime, league.CreatedBy,
	).Scan(&league.ID, &league.CreatedAt, &league.UpdatedAt)
}

// GetByID retrieves a league by ID.
func (r *LeagueRepo) GetByID(ctx context.Context, id int64) (*models.League, error) {
	query := `
		SELECT id, name, invite_code, max_members, budget_per_user, market_close_time, created_by, created_at, updated_at
		FROM leagues WHERE id = $1`

	league := &models.League{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&league.ID, &league.Name, &league.InviteCode, &league.MaxMembers,
		&league.BudgetPerUser, &league.MarketCloseTime, &league.CreatedBy,
		&league.CreatedAt, &league.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get league by id: %w", err)
	}
	return league, nil
}

// GetByInviteCode retrieves a league by invite code.
func (r *LeagueRepo) GetByInviteCode(ctx context.Context, code string) (*models.League, error) {
	query := `
		SELECT id, name, invite_code, max_members, budget_per_user, market_close_time, created_by, created_at, updated_at
		FROM leagues WHERE invite_code = $1`

	league := &models.League{}
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&league.ID, &league.Name, &league.InviteCode, &league.MaxMembers,
		&league.BudgetPerUser, &league.MarketCloseTime, &league.CreatedBy,
		&league.CreatedAt, &league.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get league by invite code: %w", err)
	}
	return league, nil
}

// GetByUserID retrieves all leagues a user belongs to.
func (r *LeagueRepo) GetByUserID(ctx context.Context, userID int64) ([]models.League, error) {
	query := `
		SELECT l.id, l.name, l.invite_code, l.max_members, l.budget_per_user,
		       l.market_close_time, l.created_by, l.created_at, l.updated_at
		FROM leagues l
		INNER JOIN league_members lm ON l.id = lm.league_id
		WHERE lm.user_id = $1
		ORDER BY l.name`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get leagues by user id: %w", err)
	}
	defer rows.Close()

	var leagues []models.League
	for rows.Next() {
		var l models.League
		err := rows.Scan(
			&l.ID, &l.Name, &l.InviteCode, &l.MaxMembers,
			&l.BudgetPerUser, &l.MarketCloseTime, &l.CreatedBy,
			&l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan league: %w", err)
		}
		leagues = append(leagues, l)
	}
	return leagues, rows.Err()
}

// Update modifies an existing league.
func (r *LeagueRepo) Update(ctx context.Context, league *models.League) error {
	query := `
		UPDATE leagues
		SET name = $1, max_members = $2, budget_per_user = $3, market_close_time = $4
		WHERE id = $5
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		league.Name, league.MaxMembers, league.BudgetPerUser,
		league.MarketCloseTime, league.ID,
	).Scan(&league.UpdatedAt)
}

// Delete removes a league.
func (r *LeagueRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM leagues WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// AddMember adds a user to a league.
func (r *LeagueRepo) AddMember(ctx context.Context, member *models.LeagueMember) error {
	query := `
		INSERT INTO league_members (league_id, user_id, role, budget)
		VALUES ($1, $2, $3, $4)
		RETURNING id, joined_at`

	return r.pool.QueryRow(ctx, query,
		member.LeagueID, member.UserID, member.Role, member.Budget,
	).Scan(&member.ID, &member.JoinedAt)
}

// GetMembers retrieves all members of a league.
func (r *LeagueRepo) GetMembers(ctx context.Context, leagueID int64) ([]models.LeagueMember, error) {
	query := `
		SELECT id, league_id, user_id, role, budget, joined_at
		FROM league_members
		WHERE league_id = $1
		ORDER BY role, joined_at`

	rows, err := r.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get league members: %w", err)
	}
	defer rows.Close()

	var members []models.LeagueMember
	for rows.Next() {
		var m models.LeagueMember
		err := rows.Scan(&m.ID, &m.LeagueID, &m.UserID, &m.Role, &m.Budget, &m.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetMembersWithUsers retrieves all members of a league joined with their
// user info (username, display name, avatar), so callers can render real
// names without an extra fetch per member.
func (r *LeagueRepo) GetMembersWithUsers(ctx context.Context, leagueID int64) ([]models.LeagueMemberWithUser, error) {
	query := `
		SELECT lm.id, lm.league_id, lm.user_id, lm.role, lm.budget, lm.joined_at,
		       u.username, u.display_name, u.avatar_url
		FROM league_members lm
		INNER JOIN users u ON u.id = lm.user_id
		WHERE lm.league_id = $1
		ORDER BY lm.role, lm.joined_at`

	rows, err := r.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get league members with users: %w", err)
	}
	defer rows.Close()

	var members []models.LeagueMemberWithUser
	for rows.Next() {
		var m models.LeagueMemberWithUser
		err := rows.Scan(
			&m.ID, &m.LeagueID, &m.UserID, &m.Role, &m.Budget, &m.JoinedAt,
			&m.Username, &m.DisplayName, &m.AvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("scan member with user: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetMember retrieves a specific member of a league.
func (r *LeagueRepo) GetMember(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error) {
	query := `
		SELECT id, league_id, user_id, role, budget, joined_at
		FROM league_members
		WHERE league_id = $1 AND user_id = $2`

	member := &models.LeagueMember{}
	err := r.pool.QueryRow(ctx, query, leagueID, userID).Scan(
		&member.ID, &member.LeagueID, &member.UserID,
		&member.Role, &member.Budget, &member.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get league member: %w", err)
	}
	return member, nil
}

// UpdateMemberBudget updates a member's remaining budget.
func (r *LeagueRepo) UpdateMemberBudget(ctx context.Context, leagueID, userID int64, budget int) error {
	query := `UPDATE league_members SET budget = $1 WHERE league_id = $2 AND user_id = $3`
	_, err := r.pool.Exec(ctx, query, budget, leagueID, userID)
	return err
}

// RemoveMember removes a user from a league.
func (r *LeagueRepo) RemoveMember(ctx context.Context, leagueID, userID int64) error {
	query := `DELETE FROM league_members WHERE league_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, leagueID, userID)
	return err
}

// CountMembers returns the number of members in a league.
func (r *LeagueRepo) CountMembers(ctx context.Context, leagueID int64) (int, error) {
	query := `SELECT COUNT(*) FROM league_members WHERE league_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, leagueID).Scan(&count)
	return count, err
}
