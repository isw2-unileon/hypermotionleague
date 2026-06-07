// Command ingest is the one-shot pipeline that turns real match data into
// standings points for one GLOBAL app matchday:
//
//	fetch per-player stats (API-Football)  →  ComputePoints (scoring engine)
//	  →  upsert player_points (global, keyed (player, matchday))
//	  →  propagate to EVERY league's lineup_players.points / lineups.total_points
//	  →  read by GetStandings.
//
// Matchdays are GLOBAL: a load creates/uses "app matchday N" (--app-matchday),
// where N is the app's OWN sequential counter (1, 2, 3...), independent of the
// real Segunda round (--round, used only to fetch from API-Football). With
// --app-matchday=0 the next number (MAX+1) is used. The computed points are
// universal and propagate to ALL leagues; each manager scores according to the
// lineup they set for that matchday — no lineup means no points that matchday.
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
	"time"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/apifootball"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/config"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/db"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/repository/postgres"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/scoring"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	round := flag.Int("round", 0, "real Segunda round to fetch from API-Football (required, e.g. 30)")
	season := flag.Int("season", 2025, "season start year (2025 = the 2025-26 season)")
	league := flag.Int("league", 141, "API-Football league ID for the fetcher (141 = Segunda División)")
	appMatchday := flag.Int("app-matchday", 0,
		"app sequential matchday number to load into (0 = next, i.e. MAX(number)+1)")
	dryRun := flag.Bool("dry-run", true,
		"fetch + compute and log what WOULD be written, making zero DB writes; pass --dry-run=false to persist")
	flag.Parse()

	if *round <= 0 {
		logger.Error("invalid flags", "error", "--round is required and must be > 0")
		os.Exit(2)
	}

	if err := run(*dryRun, *round, *season, *league, *appMatchday); err != nil {
		logger.Error("ingest failed", "error", err)
		os.Exit(1)
	}
}

func run(dryRun bool, round, season, league, appMatchday int) error {
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
		"api_league", league, "app_matchday", appMatchday,
	)

	// Resolve or create the GLOBAL app matchday. `number` is the app's own
	// sequential counter, NOT the real Segunda round. player_points and lineups
	// (across all leagues) key off this single global matchday id.
	md, created, err := resolveOrCreateMatchday(ctx, matchdayRepo, appMatchday, dryRun)
	if err != nil {
		return err
	}
	matchdayID := md.ID
	logger.Info("using app matchday",
		"number", md.Number, "matchday_id", matchdayID, "created", created)

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

	// Propagate the computed points into EVERY league's lineups for this global
	// matchday: lineup_players.points + lineups.total_points (read by
	// GetStandings). PropagateMatchdayPoints already works league-agnostically —
	// it keys on the global matchday id. Skipped on a dry run (it is a write).
	lineupsPropagated := 0
	if dryRun {
		logger.Info("would propagate points to lineups (all leagues)", "matchday_id", matchdayID)
	} else {
		lineupsPropagated, err = matchdayRepo.PropagateMatchdayPoints(ctx, matchdayID)
		if err != nil {
			return fmt.Errorf("propagate matchday points: %w", err)
		}
	}

	logger.Info("ingest complete",
		"dry_run", dryRun,
		"app_matchday", md.Number,
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

// resolveOrCreateMatchday finds the GLOBAL app matchday with the given
// sequential number, creating it if it does not exist. With appMatchday <= 0 the
// next number (MAX+1) is used. The returned bool reports whether it was created.
//
// A freshly loaded round is finished, so a created matchday gets a recent past
// window (not is_current) — this keeps the market window (which reads matchday
// dates) from treating it as "in play". Adjust the dates via the matchday
// endpoint if you need different ones. In dry-run nothing is written: a
// to-be-created matchday is returned with ID 0 (never used for writes).
func resolveOrCreateMatchday(ctx context.Context, repo *postgres.MatchdayRepo, appMatchday int, dryRun bool) (*models.Matchday, bool, error) {
	number := appMatchday
	if number <= 0 {
		all, err := repo.GetAll(ctx)
		if err != nil {
			return nil, false, fmt.Errorf("list matchdays: %w", err)
		}
		maxNum := 0
		for _, m := range all {
			if m.Number > maxNum {
				maxNum = m.Number
			}
		}
		number = maxNum + 1
	}

	existing, err := repo.GetByNumber(ctx, number)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	now := time.Now()
	md := &models.Matchday{
		Number:    number,
		Name:      fmt.Sprintf("Jornada %d", number),
		StartDate: now.Add(-2 * time.Hour), // finished round: recent past window
		EndDate:   now.Add(-1 * time.Hour),
		IsCurrent: false,
	}

	if dryRun {
		logger.Info("would create app matchday", "number", number, "name", md.Name)
		return md, true, nil // ID stays 0; never persisted in dry-run
	}

	if err := repo.Create(ctx, md); err != nil {
		return nil, false, fmt.Errorf("create app matchday %d: %w", number, err)
	}
	return md, true, nil
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
