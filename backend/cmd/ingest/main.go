// Command ingest is the one-shot pipeline that turns real match data into
// standings points for a single matchday:
//
//	fetch per-player stats (API-Football)  →  ComputePoints (scoring engine)
//	  →  upsert player_points  →  propagate to lineup_players.points and
//	  lineups.total_points  →  read by GetStandings.
//
// Defaults target La Liga Hypermotion (Spanish Segunda División, API-Football
// league 141) for the 2025-26 season (season=2025), and the seed fantasy league
// (leagues.id = 1). Note the two distinct "league" ids: --league is the
// API-Football competition used by the fetcher, while --fantasy-league is our
// internal leagues.id, which scopes the matchday row and the standings written.
//
// It is dry-run by default: with no flags it fetches + computes and logs what
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
	"syscall"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/apifootball"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/scoring"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	round := flag.Int("round", 0, "matchday number to ingest (required, e.g. 30)")
	season := flag.Int("season", 2025, "season start year (2025 = the 2025-26 season)")
	league := flag.Int("league", 141, "API-Football league ID for the fetcher (141 = Segunda División)")
	fantasyLeague := flag.Int64("fantasy-league", 1, "internal fantasy league ID (leagues.id) whose matchday/standings to write")
	dryRun := flag.Bool("dry-run", true,
		"fetch + compute and log what WOULD be written, making zero DB writes; pass --dry-run=false to persist")
	flag.Parse()

	if *round <= 0 {
		logger.Error("invalid flags", "error", "--round is required and must be > 0")
		os.Exit(2)
	}

	if err := run(*dryRun, *round, *season, *league, *fantasyLeague); err != nil {
		logger.Error("ingest failed", "error", err)
		os.Exit(1)
	}
}

func run(dryRun bool, round, season, league int, fantasyLeague int64) error {
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

	playerRepo := postgres.NewPlayerRepo(pool.Pool)
	matchdayRepo := postgres.NewMatchdayRepo(pool.Pool)
	client := apifootball.NewClient(cfg.APIFootballKey)

	logger.Info("starting ingest",
		"dry_run", dryRun, "round", round, "season", season,
		"api_league", league, "fantasy_league", fantasyLeague,
	)

	// Resolve (fantasy league, round number) -> matchday_id. player_points and
	// lineups key off this internal matchday row, not the bare round number, and
	// matchdays are league-scoped — so the same resolution the standings handler
	// uses (GetByLeague + match on Number) applies here.
	matchdayID, err := resolveMatchdayID(ctx, matchdayRepo, fantasyLeague, round)
	if err != nil {
		return err
	}
	logger.Info("resolved matchday", "fantasy_league", fantasyLeague, "round", round, "matchday_id", matchdayID)

	// Fetch per-player stats for the whole round (~12 API calls for Segunda).
	stats, err := client.FetchFixturePlayerStats(ctx, league, season, round)
	if err != nil {
		return fmt.Errorf("fetch fixture player stats: %w", err)
	}
	logger.Info("fetched player stats", "count", len(stats))

	var matched, unmatched, upserted int
	for _, s := range stats {
		player, err := playerRepo.GetByExternalID(ctx, int64(s.PlayerExternalID))
		if err != nil {
			return fmt.Errorf("resolve player external_id=%d: %w", s.PlayerExternalID, err)
		}
		if player == nil {
			// external_id not in our players table: skip but count it, so an
			// incomplete load is visible (and exits non-zero at the end).
			unmatched++
			logger.Warn("unmatched player skipped (external_id not in DB)",
				"external_id", s.PlayerExternalID, "team_external_id", s.TeamExternalID)
			continue
		}
		matched++

		pts := scoring.ComputePoints(string(player.Position), toMatchStats(s))

		pp := &models.PlayerPoints{
			PlayerID:      player.ID,
			MatchdayID:    matchdayID,
			Points:        pts,
			Goals:         s.Goals,
			Assists:       s.Assists,
			MinutesPlayed: s.Minutes,
			YellowCards:   s.Yellow,
			RedCards:      s.Red,
			CleanSheet:    s.CleanSheet,
			GoalsConceded: s.GoalsConceded,
			PensMissed:    s.PensMissed,
			PensSaved:     s.PensSaved,
			Saves:         s.Saves,
		}

		if dryRun {
			logger.Info("would upsert player_points",
				"player_id", player.ID, "external_id", s.PlayerExternalID,
				"position", player.Position, "minutes", s.Minutes, "points", pts)
			continue
		}

		if err := playerRepo.UpsertPoints(ctx, pp); err != nil {
			return fmt.Errorf("upsert points for player_id=%d: %w", player.ID, err)
		}
		upserted++
	}

	// Propagate the computed points into the matchday's lineups so the standings
	// reflect them: lineup_players.points (per-matchday) + lineups.total_points
	// (overall). Skipped on a dry run (it is a write).
	lineupsPropagated := 0
	if dryRun {
		logger.Info("would propagate points to lineups", "matchday_id", matchdayID)
	} else {
		lineupsPropagated, err = matchdayRepo.PropagateMatchdayPoints(ctx, matchdayID)
		if err != nil {
			return fmt.Errorf("propagate matchday points: %w", err)
		}
	}

	logger.Info("ingest complete",
		"dry_run", dryRun,
		"stats_fetched", len(stats),
		"players_matched", matched,
		"players_unmatched", unmatched,
		"points_upserted", upserted,
		"lineups_propagated", lineupsPropagated,
	)

	// Surface incomplete coverage as a non-zero exit so a partial load is not
	// mistaken for a clean one.
	if unmatched > 0 {
		return fmt.Errorf("incomplete coverage: %d fetched players had no matching external_id in the players table", unmatched)
	}
	return nil
}

// resolveMatchdayID maps a (fantasy league, round number) to the internal
// matchdays.id, mirroring how the standings handler resolves a matchday by
// number within a league.
func resolveMatchdayID(ctx context.Context, repo *postgres.MatchdayRepo, leagueID int64, round int) (int64, error) {
	matchdays, err := repo.GetByLeague(ctx, leagueID)
	if err != nil {
		return 0, fmt.Errorf("get matchdays for fantasy league %d: %w", leagueID, err)
	}
	for _, md := range matchdays {
		if md.Number == round {
			return md.ID, nil
		}
	}
	return 0, fmt.Errorf("no matchday number %d in fantasy league %d (run migrations/seed first?)", round, leagueID)
}

// toMatchStats maps a fetched DTO to the scoring engine's input. OwnGoals has no
// source in /fixtures/players (own goals live in /fixtures/events), so it stays
// 0; goals_conceded and clean_sheet are already team-derived by the fetcher.
func toMatchStats(s apifootball.PlayerMatchStatsDTO) scoring.PlayerMatchStats {
	return scoring.PlayerMatchStats{
		Minutes:       s.Minutes,
		Goals:         s.Goals,
		Assists:       s.Assists,
		OwnGoals:      0,
		Yellow:        s.Yellow,
		Red:           s.Red,
		GoalsConceded: s.GoalsConceded,
		CleanSheet:    s.CleanSheet,
		PensMissed:    s.PensMissed,
		PensSaved:     s.PensSaved,
		Saves:         s.Saves,
	}
}
