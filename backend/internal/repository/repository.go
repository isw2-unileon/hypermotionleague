// Repository interfaces define the contract for data access operations,
// decoupling the business logic from the underlying data storage mechanism.
//
// Instead of directly interacting with a specific database,
// the application depends on these abstractions. This allows the data layer
// implementation to be changed without affecting the core business logic.
//
// Repositories also improve testability by enabling the use of mocks or fakes,
// so business logic can be tested independently of the database.

// Package repository defines data access interfaces.
package repository

import (
	"context"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

// UserRepository handles user data persistence.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
}

// LeagueRepository handles league data persistence.
type LeagueRepository interface {
	Create(ctx context.Context, league *models.League) error
	GetByID(ctx context.Context, id int64) (*models.League, error)
	GetByInviteCode(ctx context.Context, code string) (*models.League, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.League, error)
	Update(ctx context.Context, league *models.League) error
	Delete(ctx context.Context, id int64) error

	// Member operations
	AddMember(ctx context.Context, member *models.LeagueMember) error
	GetMembers(ctx context.Context, leagueID int64) ([]models.LeagueMember, error)
	GetMember(ctx context.Context, leagueID, userID int64) (*models.LeagueMember, error)
	UpdateMemberBudget(ctx context.Context, leagueID, userID int64, budget int) error
	RemoveMember(ctx context.Context, leagueID, userID int64) error
	CountMembers(ctx context.Context, leagueID int64) (int, error)
}

// PlayerRepository handles player data persistence.
type PlayerRepository interface {
	Create(ctx context.Context, player *models.Player) error
	GetByID(ctx context.Context, id int64) (*models.Player, error)
	List(ctx context.Context, position *models.PlayerPosition, teamName *string) ([]models.Player, error)
	GetAvailableForLeague(ctx context.Context, leagueID int64) ([]models.Player, error)
	Update(ctx context.Context, player *models.Player) error
	Delete(ctx context.Context, id int64) error

	// Points operations
	UpsertPoints(ctx context.Context, points *models.PlayerPoints) error
	GetPoints(ctx context.Context, playerID, matchdayID int64) (*models.PlayerPoints, error)
	GetPointsByMatchday(ctx context.Context, matchdayID int64) ([]models.PlayerPoints, error)
}

// TeamRepository handles team/player ownership persistence.
type TeamRepository interface {
	AddPlayer(ctx context.Context, tp *models.TeamPlayer) error
	RemovePlayer(ctx context.Context, leagueID, userID, playerID int64) error
	GetUserTeam(ctx context.Context, leagueID, userID int64) (*models.UserTeam, error)
	GetPlayerOwner(ctx context.Context, leagueID, playerID int64) (*models.TeamPlayer, error)
	HasPlayer(ctx context.Context, leagueID, userID, playerID int64) (bool, error)
	TransferPlayer(ctx context.Context, leagueID, oldUserID, newUserID, playerID int64, price int) error
}

// MatchdayRepository handles matchday and lineup persistence.
type MatchdayRepository interface {
	// Matchday operations
	Create(ctx context.Context, matchday *models.Matchday) error
	GetByID(ctx context.Context, id int64) (*models.Matchday, error)
	GetByLeague(ctx context.Context, leagueID int64) ([]models.Matchday, error)
	GetCurrent(ctx context.Context, leagueID int64) (*models.Matchday, error)
	Update(ctx context.Context, matchday *models.Matchday) error

	// Lineup operations
	CreateLineup(ctx context.Context, lineup *models.Lineup) error
	GetLineup(ctx context.Context, leagueID, userID, matchdayID int64) (*models.LineupWithPlayers, error)
	UpsertLineupPlayer(ctx context.Context, lp *models.LineupPlayer) error
	RemoveLineupPlayer(ctx context.Context, lineupID, playerID int64) error
	UpdateLineupPoints(ctx context.Context, lineupID int64, totalPoints int) error

	// Standings operations
	GetStandings(ctx context.Context, leagueID int64, matchdayID *int64) (*models.Standings, error)
}

// MarketRepository handles market listings and bids persistence.
type MarketRepository interface {
	// Listing operations
	CreateListing(ctx context.Context, listing *models.MarketListing) error
	GetListingByID(ctx context.Context, id int64) (*models.MarketListingWithDetails, error)
	GetActiveListings(ctx context.Context, leagueID int64) ([]models.MarketListingWithDetails, error)
	UpdateListingStatus(ctx context.Context, id int64, status string) error
	GetExpiredListings(ctx context.Context) ([]models.MarketListing, error)

	// Bid operations
	PlaceBid(ctx context.Context, bid *models.Bid) error
	GetBidsByListing(ctx context.Context, listingID int64) ([]models.Bid, error)
	GetHighestBid(ctx context.Context, listingID int64) (*models.Bid, error)
	GetUserActiveBids(ctx context.Context, userID int64) ([]models.BidWithDetails, error)
	CountUserActiveBids(ctx context.Context, userID int64) (int, error)
	UpdateBidStatus(ctx context.Context, id int64, status models.BidStatus) error
	CancelBid(ctx context.Context, bidID, userID int64) error

	// Market status
	GetMarketStatus(ctx context.Context, leagueID, userID int64) (*models.MarketStatus, error)
}
