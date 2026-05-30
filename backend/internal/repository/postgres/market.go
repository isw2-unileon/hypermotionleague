package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MarketRepo implements repository.MarketRepository for PostgreSQL.
type MarketRepo struct {
	pool *pgxpool.Pool
}

// NewMarketRepo creates a new MarketRepo.
func NewMarketRepo(pool *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{pool: pool}
}

// CreateListing inserts a new market listing.
func (r *MarketRepo) CreateListing(ctx context.Context, listing *models.MarketListing) error {
	query := `
		INSERT INTO market_listings (league_id, player_id, base_price, seller_id, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, listed_at`

	return r.pool.QueryRow(ctx, query,
		listing.LeagueID, listing.PlayerID, listing.BasePrice,
		listing.SellerID, listing.Status, listing.ExpiresAt,
	).Scan(&listing.ID, &listing.ListedAt)
}

// GetListingByID retrieves a listing with details.
func (r *MarketRepo) GetListingByID(ctx context.Context, id int64) (*models.MarketListingWithDetails, error) {
	query := `
		SELECT ml.id, ml.league_id, ml.player_id, ml.base_price, ml.seller_id,
		       ml.status, ml.listed_at, ml.expires_at,
		       p.id, p.first_name, p.last_name, p.position, p.team_name,
		       p.market_value, p.is_active, p.created_at, p.updated_at,
		       u.display_name,
		       (SELECT MAX(b.amount) FROM bids b WHERE b.listing_id = ml.id AND b.status = 'active'),
		       (SELECT COUNT(*) FROM bids b WHERE b.listing_id = ml.id AND b.status = 'active')
		FROM market_listings ml
		INNER JOIN players p ON ml.player_id = p.id
		LEFT JOIN users u ON ml.seller_id = u.id
		WHERE ml.id = $1`

	ld := &models.MarketListingWithDetails{}
	var sellerName *string
	var highestBid *int

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ld.ID, &ld.LeagueID, &ld.PlayerID, &ld.BasePrice, &ld.SellerID,
		&ld.Status, &ld.ListedAt, &ld.ExpiresAt,
		&ld.Player.ID, &ld.Player.FirstName, &ld.Player.LastName,
		&ld.Player.Position, &ld.Player.TeamName, &ld.Player.MarketValue,
		&ld.Player.IsActive, &ld.Player.CreatedAt, &ld.Player.UpdatedAt,
		&sellerName, &highestBid, &ld.BidCount,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get listing by id: %w", err)
	}

	ld.SellerName = sellerName
	ld.HighestBid = highestBid
	return ld, nil
}

// GetActiveListings retrieves all active listings for a league.
func (r *MarketRepo) GetActiveListings(ctx context.Context, leagueID int64) ([]models.MarketListingWithDetails, error) {
	query := `
		SELECT ml.id, ml.league_id, ml.player_id, ml.base_price, ml.seller_id,
		       ml.status, ml.listed_at, ml.expires_at,
		       p.id, p.first_name, p.last_name, p.position, p.team_name,
		       p.market_value, p.is_active, p.created_at, p.updated_at,
		       u.display_name,
		       (SELECT MAX(b.amount) FROM bids b WHERE b.listing_id = ml.id AND b.status = 'active'),
		       (SELECT COUNT(*) FROM bids b WHERE b.listing_id = ml.id AND b.status = 'active')
		FROM market_listings ml
		INNER JOIN players p ON ml.player_id = p.id
		LEFT JOIN users u ON ml.seller_id = u.id
		WHERE ml.league_id = $1 AND ml.status = 'active'
		ORDER BY ml.expires_at, p.market_value DESC`

	rows, err := r.pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get active listings: %w", err)
	}
	defer rows.Close()

	var listings []models.MarketListingWithDetails
	for rows.Next() {
		var ld models.MarketListingWithDetails
		var sellerName *string
		var highestBid *int

		err := rows.Scan(
			&ld.ID, &ld.LeagueID, &ld.PlayerID, &ld.BasePrice, &ld.SellerID,
			&ld.Status, &ld.ListedAt, &ld.ExpiresAt,
			&ld.Player.ID, &ld.Player.FirstName, &ld.Player.LastName,
			&ld.Player.Position, &ld.Player.TeamName, &ld.Player.MarketValue,
			&ld.Player.IsActive, &ld.Player.CreatedAt, &ld.Player.UpdatedAt,
			&sellerName, &highestBid, &ld.BidCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan listing: %w", err)
		}
		ld.SellerName = sellerName
		ld.HighestBid = highestBid
		listings = append(listings, ld)
	}
	return listings, rows.Err()
}

// UpdateListingStatus updates the status of a listing.
func (r *MarketRepo) UpdateListingStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE market_listings SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

// GetExpiredListings retrieves all listings that have expired and are still active.
func (r *MarketRepo) GetExpiredListings(ctx context.Context) ([]models.MarketListing, error) {
	query := `
		SELECT id, league_id, player_id, base_price, seller_id, status, listed_at, expires_at
		FROM market_listings
		WHERE status = 'active' AND expires_at <= NOW()
		ORDER BY expires_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get expired listings: %w", err)
	}
	defer rows.Close()

	var listings []models.MarketListing
	for rows.Next() {
		var l models.MarketListing
		err := rows.Scan(
			&l.ID, &l.LeagueID, &l.PlayerID, &l.BasePrice,
			&l.SellerID, &l.Status, &l.ListedAt, &l.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan listing: %w", err)
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

// PlaceBid inserts a new bid.
func (r *MarketRepo) PlaceBid(ctx context.Context, bid *models.Bid) error {
	query := `
		INSERT INTO bids (listing_id, user_id, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, placed_at`

	return r.pool.QueryRow(ctx, query,
		bid.ListingID, bid.UserID, bid.Amount, bid.Status,
	).Scan(&bid.ID, &bid.PlacedAt)
}

// GetBidsByListing retrieves all bids for a listing.
func (r *MarketRepo) GetBidsByListing(ctx context.Context, listingID int64) ([]models.Bid, error) {
	query := `
		SELECT id, listing_id, user_id, amount, status, placed_at
		FROM bids
		WHERE listing_id = $1
		ORDER BY amount DESC, placed_at ASC`

	rows, err := r.pool.Query(ctx, query, listingID)
	if err != nil {
		return nil, fmt.Errorf("get bids by listing: %w", err)
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var b models.Bid
		err := rows.Scan(&b.ID, &b.ListingID, &b.UserID, &b.Amount, &b.Status, &b.PlacedAt)
		if err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}
		bids = append(bids, b)
	}
	return bids, rows.Err()
}

// GetHighestBid retrieves the highest active bid for a listing.
func (r *MarketRepo) GetHighestBid(ctx context.Context, listingID int64) (*models.Bid, error) {
	query := `
		SELECT id, listing_id, user_id, amount, status, placed_at
		FROM bids
		WHERE listing_id = $1 AND status = 'active'
		ORDER BY amount DESC, placed_at ASC
		LIMIT 1`

	bid := &models.Bid{}
	err := r.pool.QueryRow(ctx, query, listingID).Scan(
		&bid.ID, &bid.ListingID, &bid.UserID, &bid.Amount, &bid.Status, &bid.PlacedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get highest bid: %w", err)
	}
	return bid, nil
}

// GetUserActiveBids retrieves all active bids for a user with details.
func (r *MarketRepo) GetUserActiveBids(ctx context.Context, userID int64) ([]models.BidWithDetails, error) {
	query := `
		SELECT b.id, b.listing_id, b.user_id, b.amount, b.status, b.placed_at,
		       ml.id, ml.league_id, ml.player_id, ml.base_price, ml.seller_id,
		       ml.status, ml.listed_at, ml.expires_at,
		       p.id, p.first_name, p.last_name, p.position, p.team_name,
		       p.market_value, p.is_active, p.created_at, p.updated_at
		FROM bids b
		INNER JOIN market_listings ml ON b.listing_id = ml.id
		INNER JOIN players p ON ml.player_id = p.id
		WHERE b.user_id = $1 AND b.status = 'active'
		ORDER BY b.placed_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user active bids: %w", err)
	}
	defer rows.Close()

	var bids []models.BidWithDetails
	for rows.Next() {
		var bwd models.BidWithDetails
		err := rows.Scan(
			&bwd.ID, &bwd.ListingID, &bwd.UserID, &bwd.Amount, &bwd.Status, &bwd.PlacedAt,
			&bwd.Listing.ID, &bwd.Listing.LeagueID, &bwd.Listing.PlayerID,
			&bwd.Listing.BasePrice, &bwd.Listing.SellerID, &bwd.Listing.Status,
			&bwd.Listing.ListedAt, &bwd.Listing.ExpiresAt,
			&bwd.Player.ID, &bwd.Player.FirstName, &bwd.Player.LastName,
			&bwd.Player.Position, &bwd.Player.TeamName, &bwd.Player.MarketValue,
			&bwd.Player.IsActive, &bwd.Player.CreatedAt, &bwd.Player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bid with details: %w", err)
		}
		bids = append(bids, bwd)
	}
	return bids, rows.Err()
}

// CountUserActiveBids counts active bids for a user.
func (r *MarketRepo) CountUserActiveBids(ctx context.Context, userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM bids WHERE user_id = $1 AND status = 'active'`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// UpdateBidStatus updates the status of a bid.
func (r *MarketRepo) UpdateBidStatus(ctx context.Context, id int64, status models.BidStatus) error {
	query := `UPDATE bids SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

// CancelBid cancels a user's bid.
func (r *MarketRepo) CancelBid(ctx context.Context, bidID, userID int64) error {
	query := `UPDATE bids SET status = 'cancelled' WHERE id = $1 AND user_id = $2 AND status = 'active'`
	result, err := r.pool.Exec(ctx, query, bidID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("bid not found or already processed")
	}
	return nil
}

// GetMarketStatus retrieves the market status for a user in a league.
func (r *MarketRepo) GetMarketStatus(ctx context.Context, leagueID, userID int64) (*models.MarketStatus, error) {
	query := `
		SELECT l.market_close_time,
		       (SELECT COUNT(*) FROM market_listings ml WHERE ml.league_id = $1 AND ml.status = 'active'),
		       (SELECT COUNT(*) FROM bids b
		        INNER JOIN market_listings ml ON b.listing_id = ml.id
		        WHERE ml.league_id = $1 AND b.user_id = $2 AND b.status = 'active')
		FROM leagues l
		WHERE l.id = $1`

	var closeTime string
	status := &models.MarketStatus{LeagueID: leagueID, MaxBidsPerUser: 5}

	err := r.pool.QueryRow(ctx, query, leagueID, userID).Scan(
		&closeTime, &status.ActiveListings, &status.YourActiveBids,
	)
	if err != nil {
		return nil, fmt.Errorf("get market status: %w", err)
	}

	return status, nil
}

// PlaceBidTx inserts a new bid inside a transaction, serializing concurrent requests
// and properly calculating the committed budget to prevent overdrafts.
func (r *MarketRepo) PlaceBidTx(ctx context.Context, leagueID int64, bid *models.Bid) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// Si algo falla a mitad de camino, revierte todo (rollback)
	defer tx.Rollback(ctx)

	// 1. Bloqueo de concurrencia: SELECT FOR UPDATE
	// Congelamos la fila de este usuario en esta liga para que ninguna otra puja concurrente la lea
	var budget int
	err = tx.QueryRow(ctx, `
		SELECT budget FROM league_members 
		WHERE league_id = $1 AND user_id = $2 
		FOR UPDATE
	`, leagueID, bid.UserID).Scan(&budget)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("user is not a member of this league")
		}
		return fmt.Errorf("lock member row: %w", err)
	}

	// 2. Contar pujas activas y sumar el dinero comprometido de forma atómica
	var activeBidsCount int
	var committedBudget int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(b.amount), 0)
		FROM bids b
		JOIN market_listings ml ON b.listing_id = ml.id
		WHERE b.user_id = $1 AND ml.league_id = $2 AND b.status = 'active'
	`, bid.UserID, leagueID).Scan(&activeBidsCount, &committedBudget)
	if err != nil {
		return fmt.Errorf("calculate committed budget: %w", err)
	}

	// 3. Validaciones de negocio dentro de la transacción
	if activeBidsCount >= 5 {
		return errors.New("MAX_BIDS_REACHED")
	}

	if committedBudget+bid.Amount > budget {
		return errors.New("INSUFFICIENT_BUDGET")
	}

	// 4. Si todo es correcto, guardamos la puja
	query := `
		INSERT INTO bids (listing_id, user_id, amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, placed_at`

	err = tx.QueryRow(ctx, query, bid.ListingID, bid.UserID, bid.Amount, bid.Status).Scan(&bid.ID, &bid.PlacedAt)
	if err != nil {
		return fmt.Errorf("insert bid: %w", err)
	}

	// 5. Confirmar transacción
	return tx.Commit(ctx)
}
