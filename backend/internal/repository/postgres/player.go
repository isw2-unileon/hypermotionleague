package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlayerRepo implements repository.PlayerRepository for PostgreSQL.
type PlayerRepo struct {
	pool *pgxpool.Pool
}

// NewPlayerRepo creates a new PlayerRepo.
func NewPlayerRepo(pool *pgxpool.Pool) *PlayerRepo {
	return &PlayerRepo{pool: pool}
}

// Create inserts a new player.
func (r *PlayerRepo) Create(ctx context.Context, player *models.Player) error {
	query := `
		INSERT INTO players (first_name, last_name, position, team_name, market_value, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		player.FirstName, player.LastName, player.Position,
		player.TeamName, player.MarketValue, player.IsActive,
	).Scan(&player.ID, &player.CreatedAt, &player.UpdatedAt)
}

// GetByID retrieves a player by ID.
func (r *PlayerRepo) GetByID(ctx context.Context, id int64) (*models.Player, error) {
	query := `
		SELECT id, first_name, last_name, position, team_name, market_value, is_active, created_at, updated_at
		FROM players WHERE id = $1`

	player := &models.Player{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&player.ID, &player.FirstName, &player.LastName, &player.Position,
		&player.TeamName, &player.MarketValue, &player.IsActive,
		&player.CreatedAt, &player.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get player by id: %w", err)
	}
	return player, nil
}

// List retrieves players with optional filters.
func (r *PlayerRepo) List(ctx context.Context, position *models.PlayerPosition, teamName *string) ([]models.Player, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if position != nil {
		conditions = append(conditions, fmt.Sprintf("position = $%d", argIdx))
		args = append(args, *position)
		argIdx++
	}
	if teamName != nil {
		conditions = append(conditions, fmt.Sprintf("team_name = $%d", argIdx))
		args = append(args, *teamName)
		argIdx++
	}

	query := `SELECT id, first_name, last_name, position, team_name, market_value, is_active, created_at, updated_at FROM players`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY last_name, first_name"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var p models.Player
		err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.Position,
			&p.TeamName, &p.MarketValue, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// GetAvailableForLeague retrieves players not owned in a specific league.
func (r *PlayerRepo) GetAvailableForLeague(ctx context.Context, leagueID int64) ([]models.Player, error) {
	query := `
		SELECT p.id, p.first_name, p.last_name, p.position, p.team_name,
		       p.market_value, p.is_active, p.created_at, p.updated_at
		FROM players p
		WHERE p.is_active = TRUE
		  AND p.id NOT IN (
		      SELECT tp.player_id FROM team_players tp WHERE tp.league_id = $1
		  )
		ORDER BY p.market_value DESC`

	rows, err := r.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get available players: %w", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var p models.Player
		err := rows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &p.Position,
			&p.TeamName, &p.MarketValue, &p.IsActive,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// Update modifies an existing player.
func (r *PlayerRepo) Update(ctx context.Context, player *models.Player) error {
	query := `
		UPDATE players
		SET first_name = $1, last_name = $2, position = $3, team_name = $4,
		    market_value = $5, is_active = $6
		WHERE id = $7
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query,
		player.FirstName, player.LastName, player.Position, player.TeamName,
		player.MarketValue, player.IsActive, player.ID,
	).Scan(&player.UpdatedAt)
}

// Delete removes a player.
func (r *PlayerRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM players WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpsertPoints inserts or updates player points for a matchday.
func (r *PlayerRepo) UpsertPoints(ctx context.Context, pp *models.PlayerPoints) error {
	query := `
		INSERT INTO player_points (player_id, matchday_id, points, goals, assists,
		                           minutes_played, yellow_cards, red_cards, clean_sheet)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (player_id, matchday_id)
		DO UPDATE SET
			points = EXCLUDED.points,
			goals = EXCLUDED.goals,
			assists = EXCLUDED.assists,
			minutes_played = EXCLUDED.minutes_played,
			yellow_cards = EXCLUDED.yellow_cards,
			red_cards = EXCLUDED.red_cards,
			clean_sheet = EXCLUDED.clean_sheet,
			updated_at = NOW()
		RETURNING id, updated_at`

	return r.pool.QueryRow(ctx, query,
		pp.PlayerID, pp.MatchdayID, pp.Points, pp.Goals, pp.Assists,
		pp.MinutesPlayed, pp.YellowCards, pp.RedCards, pp.CleanSheet,
	).Scan(&pp.ID, &pp.UpdatedAt)
}

// GetPoints retrieves player points for a specific matchday.
func (r *PlayerRepo) GetPoints(ctx context.Context, playerID, matchdayID int64) (*models.PlayerPoints, error) {
	query := `
		SELECT id, player_id, matchday_id, points, goals, assists,
		       minutes_played, yellow_cards, red_cards, clean_sheet, updated_at
		FROM player_points
		WHERE player_id = $1 AND matchday_id = $2`

	pp := &models.PlayerPoints{}
	err := r.pool.QueryRow(ctx, query, playerID, matchdayID).Scan(
		&pp.ID, &pp.PlayerID, &pp.MatchdayID, &pp.Points, &pp.Goals, &pp.Assists,
		&pp.MinutesPlayed, &pp.YellowCards, &pp.RedCards, &pp.CleanSheet, &pp.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get player points: %w", err)
	}
	return pp, nil
}

// GetPointsByMatchday retrieves all player points for a matchday.
func (r *PlayerRepo) GetPointsByMatchday(ctx context.Context, matchdayID int64) ([]models.PlayerPoints, error) {
	query := `
		SELECT id, player_id, matchday_id, points, goals, assists,
		       minutes_played, yellow_cards, red_cards, clean_sheet, updated_at
		FROM player_points
		WHERE matchday_id = $1
		ORDER BY points DESC`

	rows, err := r.pool.Query(ctx, query, matchdayID)
	if err != nil {
		return nil, fmt.Errorf("get points by matchday: %w", err)
	}
	defer rows.Close()

	var points []models.PlayerPoints
	for rows.Next() {
		var pp models.PlayerPoints
		err := rows.Scan(
			&pp.ID, &pp.PlayerID, &pp.MatchdayID, &pp.Points, &pp.Goals, &pp.Assists,
			&pp.MinutesPlayed, &pp.YellowCards, &pp.RedCards, &pp.CleanSheet, &pp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan player points: %w", err)
		}
		points = append(points, pp)
	}
	return points, rows.Err()
}
