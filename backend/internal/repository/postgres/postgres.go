// Package postgres implements repository interfaces using PostgreSQL.
package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories holds all PostgreSQL repository implementations.
type Repositories struct {
	User     *UserRepo
	League   *LeagueRepo
	Player   *PlayerRepo
	Team     *TeamRepo
	Matchday *MatchdayRepo
	Market   *MarketRepo
}

// NewRepositories creates all repository implementations.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:     NewUserRepo(pool),
		League:   NewLeagueRepo(pool),
		Player:   NewPlayerRepo(pool),
		Team:     NewTeamRepo(pool),
		Matchday: NewMatchdayRepo(pool),
		Market:   NewMarketRepo(pool),
	}
}
