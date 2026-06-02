package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClubRepo implements repository.ClubRepository for PostgreSQL.
//
// It manages the real-club catalog (table `teams`), which is unrelated to the
// fantasy-ownership TeamRepo (table `team_players`).
type ClubRepo struct {
	pool *pgxpool.Pool
}

// NewClubRepo creates a new ClubRepo.
func NewClubRepo(pool *pgxpool.Pool) *ClubRepo {
	return &ClubRepo{pool: pool}
}

// UpsertByExternalID inserts a club or updates it on external_id conflict,
// returning the internal id. founded is nullable (a nil *int writes SQL NULL).
func (r *ClubRepo) UpsertByExternalID(ctx context.Context, club *models.Team) (int64, error) {
	query := `
		INSERT INTO teams (external_id, name, code, country, founded, logo_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (external_id) DO UPDATE SET
			name = EXCLUDED.name,
			code = EXCLUDED.code,
			country = EXCLUDED.country,
			founded = EXCLUDED.founded,
			logo_url = EXCLUDED.logo_url,
			updated_at = NOW()
		RETURNING id`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		club.ExternalID, club.Name, club.Code, club.Country, club.Founded, club.LogoURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert club by external id %d: %w", club.ExternalID, err)
	}
	return id, nil
}

// GetByExternalID retrieves a club by its API-Football external ID.
// COALESCE guards the nullable text columns so they scan into plain strings.
func (r *ClubRepo) GetByExternalID(ctx context.Context, externalID int64) (*models.Team, error) {
	query := `
		SELECT id, external_id, name,
		       COALESCE(code, ''), COALESCE(country, ''), founded, COALESCE(logo_url, ''),
		       created_at, updated_at
		FROM teams WHERE external_id = $1`

	club := &models.Team{}
	err := r.pool.QueryRow(ctx, query, externalID).Scan(
		&club.ID, &club.ExternalID, &club.Name,
		&club.Code, &club.Country, &club.Founded, &club.LogoURL,
		&club.CreatedAt, &club.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get club by external id %d: %w", externalID, err)
	}
	return club, nil
}
