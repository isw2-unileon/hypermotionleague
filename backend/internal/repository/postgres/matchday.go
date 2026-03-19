package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MatchdayRepo implements repository.MatchdayRepository for PostgreSQL.
type MatchdayRepo struct {
	pool *pgxpool.Pool
}

// NewMatchdayRepo creates a new MatchdayRepo.
func NewMatchdayRepo(pool *pgxpool.Pool) *MatchdayRepo {
	return &MatchdayRepo{pool: pool}
}

// Create inserts a new matchday.
func (r *MatchdayRepo) Create(ctx context.Context, matchday *models.Matchday) error {
	query := `
		INSERT INTO matchdays (league_id, number, name, start_date, end_date, is_current)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		matchday.LeagueID, matchday.Number, matchday.Name,
		matchday.StartDate, matchday.EndDate, matchday.IsCurrent,
	).Scan(&matchday.ID, &matchday.CreatedAt)
}

// GetByID retrieves a matchday by ID.
func (r *MatchdayRepo) GetByID(ctx context.Context, id int64) (*models.Matchday, error) {
	query := `
		SELECT id, league_id, number, name, start_date, end_date, is_current, created_at
		FROM matchdays WHERE id = $1`

	m := &models.Matchday{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.LeagueID, &m.Number, &m.Name,
		&m.StartDate, &m.EndDate, &m.IsCurrent, &m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get matchday by id: %w", err)
	}
	return m, nil
}

// GetByLeague retrieves all matchdays for a league.
func (r *MatchdayRepo) GetByLeague(ctx context.Context, leagueID int64) ([]models.Matchday, error) {
	query := `
		SELECT id, league_id, number, name, start_date, end_date, is_current, created_at
		FROM matchdays
		WHERE league_id = $1
		ORDER BY number`

	rows, err := r.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get matchdays by league: %w", err)
	}
	defer rows.Close()

	var matchdays []models.Matchday
	for rows.Next() {
		var m models.Matchday
		err := rows.Scan(
			&m.ID, &m.LeagueID, &m.Number, &m.Name,
			&m.StartDate, &m.EndDate, &m.IsCurrent, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan matchday: %w", err)
		}
		matchdays = append(matchdays, m)
	}
	return matchdays, rows.Err()
}

// GetCurrent retrieves the current matchday for a league.
func (r *MatchdayRepo) GetCurrent(ctx context.Context, leagueID int64) (*models.Matchday, error) {
	query := `
		SELECT id, league_id, number, name, start_date, end_date, is_current, created_at
		FROM matchdays
		WHERE league_id = $1 AND is_current = TRUE
		LIMIT 1`

	m := &models.Matchday{}
	err := r.pool.QueryRow(ctx, query, leagueID).Scan(
		&m.ID, &m.LeagueID, &m.Number, &m.Name,
		&m.StartDate, &m.EndDate, &m.IsCurrent, &m.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get current matchday: %w", err)
	}
	return m, nil
}

// Update modifies an existing matchday.
func (r *MatchdayRepo) Update(ctx context.Context, matchday *models.Matchday) error {
	query := `
		UPDATE matchdays
		SET name = $1, start_date = $2, end_date = $3, is_current = $4
		WHERE id = $5`

	_, err := r.pool.Exec(ctx, query,
		matchday.Name, matchday.StartDate, matchday.EndDate,
		matchday.IsCurrent, matchday.ID,
	)
	return err
}

// CreateLineup creates a new lineup.
func (r *MatchdayRepo) CreateLineup(ctx context.Context, lineup *models.Lineup) error {
	query := `
		INSERT INTO lineups (league_id, user_id, matchday_id, total_points)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		lineup.LeagueID, lineup.UserID, lineup.MatchdayID, lineup.TotalPoints,
	).Scan(&lineup.ID, &lineup.CreatedAt, &lineup.UpdatedAt)
}

// GetLineup retrieves a user's lineup for a matchday with player details.
func (r *MatchdayRepo) GetLineup(ctx context.Context, leagueID, userID, matchdayID int64) (*models.LineupWithPlayers, error) {
	lineupQuery := `
		SELECT id, league_id, user_id, matchday_id, total_points, created_at, updated_at
		FROM lineups
		WHERE league_id = $1 AND user_id = $2 AND matchday_id = $3`

	lineup := &models.LineupWithPlayers{}
	err := r.pool.QueryRow(ctx, lineupQuery, leagueID, userID, matchdayID).Scan(
		&lineup.ID, &lineup.LeagueID, &lineup.UserID, &lineup.MatchdayID,
		&lineup.TotalPoints, &lineup.CreatedAt, &lineup.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get lineup: %w", err)
	}

	// Get lineup players
	playerQuery := `
		SELECT lp.id, lp.lineup_id, lp.player_id, lp.position, lp.is_starter, lp.points,
		       p.id, p.first_name, p.last_name, p.position, p.team_name, p.market_value,
		       p.is_active, p.created_at, p.updated_at
		FROM lineup_players lp
		INNER JOIN players p ON lp.player_id = p.id
		WHERE lp.lineup_id = $1
		ORDER BY lp.is_starter DESC, lp.position`

	rows, err := r.pool.Query(ctx, playerQuery, lineup.ID)
	if err != nil {
		return nil, fmt.Errorf("get lineup players: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var lpd models.LineupPlayerWithDetails
		err := rows.Scan(
			&lpd.ID, &lpd.LineupID, &lpd.PlayerID, &lpd.Position,
			&lpd.IsStarter, &lpd.Points,
			&lpd.Player.ID, &lpd.Player.FirstName, &lpd.Player.LastName,
			&lpd.Player.Position, &lpd.Player.TeamName, &lpd.Player.MarketValue,
			&lpd.Player.IsActive, &lpd.Player.CreatedAt, &lpd.Player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan lineup player: %w", err)
		}
		lineup.Players = append(lineup.Players, lpd)
	}

	return lineup, rows.Err()
}

// UpsertLineupPlayer adds or updates a player in a lineup.
func (r *MatchdayRepo) UpsertLineupPlayer(ctx context.Context, lp *models.LineupPlayer) error {
	query := `
		INSERT INTO lineup_players (lineup_id, player_id, position, is_starter, points)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (lineup_id, player_id)
		DO UPDATE SET position = EXCLUDED.position, is_starter = EXCLUDED.is_starter
		RETURNING id`

	return r.pool.QueryRow(ctx, query,
		lp.LineupID, lp.PlayerID, lp.Position, lp.IsStarter, lp.Points,
	).Scan(&lp.ID)
}

// RemoveLineupPlayer removes a player from a lineup.
func (r *MatchdayRepo) RemoveLineupPlayer(ctx context.Context, lineupID, playerID int64) error {
	query := `DELETE FROM lineup_players WHERE lineup_id = $1 AND player_id = $2`
	_, err := r.pool.Exec(ctx, query, lineupID, playerID)
	return err
}

// UpdateLineupPoints updates the total points for a lineup.
func (r *MatchdayRepo) UpdateLineupPoints(ctx context.Context, lineupID int64, totalPoints int) error {
	query := `UPDATE lineups SET total_points = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, totalPoints, lineupID)
	return err
}

// GetStandings retrieves the standings for a league, optionally filtered by matchday.
func (r *MatchdayRepo) GetStandings(ctx context.Context, leagueID int64, matchdayID *int64) (*models.Standings, error) {
	standings := &models.Standings{LeagueID: leagueID, MatchdayID: matchdayID}

	var query string
	var args []interface{}

	if matchdayID != nil {
		// Standings for a specific matchday
		query = `
			SELECT u.id, u.username, u.display_name, COALESCE(SUM(lp.points), 0) as total_points
			FROM league_members lm
			INNER JOIN users u ON lm.user_id = u.id
			LEFT JOIN lineups l ON l.league_id = lm.league_id AND l.user_id = lm.user_id AND l.matchday_id = $2
			LEFT JOIN lineup_players lp ON lp.lineup_id = l.id
			WHERE lm.league_id = $1
			GROUP BY u.id, u.username, u.display_name
			ORDER BY total_points DESC`
		args = []interface{}{leagueID, *matchdayID}
	} else {
		// Overall standings
		query = `
			SELECT u.id, u.username, u.display_name, COALESCE(SUM(l.total_points), 0) as total_points
			FROM league_members lm
			INNER JOIN users u ON lm.user_id = u.id
			LEFT JOIN lineups l ON l.league_id = lm.league_id AND l.user_id = lm.user_id
			WHERE lm.league_id = $1
			GROUP BY u.id, u.username, u.display_name
			ORDER BY total_points DESC`
		args = []interface{}{leagueID}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get standings: %w", err)
	}
	defer rows.Close()

	rank := 0
	for rows.Next() {
		rank++
		var s models.UserStanding
		err := rows.Scan(&s.UserID, &s.Username, &s.DisplayName, &s.TotalPoints)
		if err != nil {
			return nil, fmt.Errorf("scan standing: %w", err)
		}
		s.Rank = rank
		standings.Rankings = append(standings.Rankings, s)
	}

	return standings, rows.Err()
}
