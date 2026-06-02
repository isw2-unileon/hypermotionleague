// Command sync-players loads all teams and players of a league/season from
// API-Football into Postgres, keyed on API-Football's external IDs so repeated
// runs upsert rather than duplicate.
//
// Defaults target La Liga Hypermotion (Spanish Segunda División, league 141)
// for the 2025-26 season (season=2025).
//
// It is dry-run by default: with no flags it fetches from the API and logs what
// WOULD be written, making zero DB writes. Pass --dry-run=false to persist.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/apifootball"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
)

// politenessDelay is waited between successive API calls. The API plan allows
// 7500/day so this is courtesy, not a hard requirement.
const politenessDelay = 200 * time.Millisecond

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	dryRun := flag.Bool("dry-run", true,
		"fetch from the API and log what WOULD be written, making zero DB writes; pass --dry-run=false to actually persist")
	season := flag.Int("season", 2025, "season start year (2025 = the 2025-26 season)")
	league := flag.Int("league", 141, "API-Football league ID (141 = Segunda División)")
	flag.Parse()

	if err := run(*dryRun, *season, *league); err != nil {
		logger.Error("sync-players failed", "error", err)
		os.Exit(1)
	}
}

func run(dryRun bool, season, league int) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if cfg.APIFootballKey == "" {
		return errors.New("API_FOOTBALL_KEY is not set (add it to .env)")
	}

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	repos := postgres.NewRepositories(pool.Pool)
	client := apifootball.NewClient(cfg.APIFootballKey)

	logger.Info("starting sync-players", "dry_run", dryRun, "league", league, "season", season)

	apiCalls := 0

	// --- Teams: one unpaginated call ---
	teams, err := client.GetTeams(ctx, league, season)
	apiCalls++
	if err != nil {
		return fmt.Errorf("get teams: %w", err)
	}
	logger.Info("fetched teams", "count", len(teams))

	internalByExternal := make(map[int]int64, len(teams))
	teamsUpserted := 0
	for _, t := range teams {
		club := toClub(t)
		if dryRun {
			logger.Info("would upsert team", "external_id", t.ID, "name", t.Name, "code", t.Code)
			continue
		}
		id, err := repos.Club.UpsertByExternalID(ctx, &club)
		if err != nil {
			return fmt.Errorf("upsert team %q (external_id=%d): %w", t.Name, t.ID, err)
		}
		internalByExternal[t.ID] = id
		teamsUpserted++
		logger.Info("upserted team", "external_id", t.ID, "internal_id", id, "name", t.Name)
	}

	// --- Squads: one unpaginated call per team, 200ms apart ---
	var (
		totalPlayersFetched  int
		totalPlayersUpserted int
		totalPlayersSkipped  int
		playerUpsertErrors   int
		failedSquads         []string
	)

	for _, t := range teams {
		// Politeness delay between API calls (cancellable).
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(politenessDelay):
		}

		players, err := client.GetSquad(ctx, t.ID)
		apiCalls++
		if err != nil {
			logger.Error("squad fetch failed", "team", t.Name, "external_id", t.ID, "error", err)
			failedSquads = append(failedSquads, t.Name)
			continue
		}

		fetched := len(players)
		totalPlayersFetched += fetched

		// Resolve the internal team PK once per squad (nil in dry-run).
		var teamPK *int64
		if id, ok := internalByExternal[t.ID]; ok {
			pk := id
			teamPK = &pk
		}

		upserted, skipped, upErrs := 0, 0, 0
		for _, sp := range players {
			pos, ok := mapPosition(sp.Position)
			if !ok {
				logger.Warn("unknown position, skipping player",
					"player", sp.Name, "external_id", sp.ID, "position", sp.Position, "team", t.Name)
				skipped++
				continue
			}

			player := toPlayer(sp, pos, t.Name, teamPK)
			if dryRun {
				logger.Info("would upsert player",
					"external_id", sp.ID, "name", sp.Name, "position", string(pos),
					"number", sp.Number, "age", sp.Age, "team", t.Name)
				continue
			}

			if err := repos.Player.UpsertByExternalID(ctx, &player); err != nil {
				logger.Error("player upsert failed",
					"player", sp.Name, "external_id", sp.ID, "team", t.Name, "error", err)
				upErrs++
				continue
			}
			upserted++
		}

		totalPlayersUpserted += upserted
		totalPlayersSkipped += skipped
		playerUpsertErrors += upErrs

		logger.Info("processed squad",
			"team", t.Name, "fetched", fetched, "upserted", upserted,
			"skipped", skipped, "upsert_errors", upErrs, "dry_run", dryRun)
	}

	// --- Final tally ---
	logger.Info("sync-players summary",
		"dry_run", dryRun,
		"api_calls", apiCalls,
		"teams_fetched", len(teams),
		"teams_upserted", teamsUpserted,
		"players_fetched", totalPlayersFetched,
		"players_upserted", totalPlayersUpserted,
		"players_skipped_unknown_position", totalPlayersSkipped,
		"player_upsert_errors", playerUpsertErrors,
		"failed_squads", failedSquads,
	)

	if len(failedSquads) > 0 || playerUpsertErrors > 0 {
		return fmt.Errorf("completed with errors: %d squad(s) failed %v, %d player upsert error(s)",
			len(failedSquads), failedSquads, playerUpsertErrors)
	}
	return nil
}

// mapPosition converts an API-Football position word to the repo's
// GK/DEF/MID/FWD convention. ok is false for an unrecognised value, in which
// case the caller skips the player rather than inserting an invalid enum.
func mapPosition(apiPos string) (models.PlayerPosition, bool) {
	switch strings.TrimSpace(apiPos) {
	case "Goalkeeper":
		return models.PositionGK, true
	case "Defender":
		return models.PositionDEF, true
	case "Midfielder":
		return models.PositionMID, true
	case "Attacker":
		return models.PositionFWD, true
	default:
		return "", false
	}
}

// splitName splits a combined "First Last ..." name on the first space: the
// first token becomes the first name, the remainder the last name. A
// single-token (mononym) name goes entirely into last_name so it still sorts
// naturally in the last_name-ordered player lists.
func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	if before, after, found := strings.Cut(full, " "); found {
		return before, strings.TrimSpace(after)
	}
	return "", full
}

// nullableInt returns nil for 0 (API-Football null / unknown) so the column is
// stored as SQL NULL rather than a literal 0.
func nullableInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

// toClub maps an API-Football team DTO to the real-club model.
func toClub(t apifootball.Team) models.Team {
	return models.Team{
		ExternalID: int64(t.ID),
		Name:       t.Name,
		Code:       t.Code,
		Country:    t.Country,
		Founded:    nullableInt(t.Founded),
		LogoURL:    t.Logo,
	}
}

// toPlayer maps an API-Football squad player DTO to the player model.
func toPlayer(sp apifootball.SquadPlayer, pos models.PlayerPosition, teamName string, teamPK *int64) models.Player {
	first, last := splitName(sp.Name)
	return models.Player{
		FirstName:    first,
		LastName:     last,
		Position:     pos,
		TeamName:     teamName,
		IsActive:     true,
		ExternalID:   int64(sp.ID),
		TeamID:       teamPK,
		JerseyNumber: nullableInt(sp.Number),
		PhotoURL:     sp.Photo,
		Age:          nullableInt(sp.Age),
	}
}
