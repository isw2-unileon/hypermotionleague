// Script to import real football data from API-Football into our Supabase database.
// Imports teams, players, and fixtures for La Liga Hypermotion (league 141).
//
// Usage:
//   go run backend/scripts/import_data.go
//
// Requires in .env:
//   DATABASE_URL=...
//   API_FOOTBALL_KEY=your-api-key

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	baseURL  = "https://v3.football.api-sports.io"
	leagueID = 141
	season   = 2024
)

var (
	apiKey string
	pool   *pgxpool.Pool
)

// ---------- API-Football response types ----------

type apiResponse[T any] struct {
	Response []T    `json:"response"`
	Paging   paging `json:"paging"`
	Errors   any    `json:"errors"`
}

type paging struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}

// Teams
type teamResponse struct {
	Team  teamData  `json:"team"`
	Venue venueData `json:"venue"`
}
type teamData struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Logo    string `json:"logo"`
	Country string `json:"country"`
	Founded int    `json:"founded"`
}
type venueData struct {
	Name string `json:"name"`
	City string `json:"city"`
}

// Players
type playerResponse struct {
	Player     playerData     `json:"player"`
	Statistics []playerStats  `json:"statistics"`
}
type playerData struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	Age         int    `json:"age"`
	Nationality string `json:"nationality"`
	Photo       string `json:"photo"`
}
type playerStats struct {
	Team  playerTeam     `json:"team"`
	Games playerGames    `json:"games"`
}
type playerTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type playerGames struct {
	Position string `json:"position"`
}

// Fixtures
type fixtureResponse struct {
	Fixture fixtureData    `json:"fixture"`
	League  fixtureLeague  `json:"league"`
	Teams   fixtureTeams   `json:"teams"`
	Goals   fixtureGoals   `json:"goals"`
}
type fixtureData struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	Status    struct {
		Short string `json:"short"`
	} `json:"status"`
	Venue struct {
		Name string `json:"name"`
		City string `json:"city"`
	} `json:"venue"`
}
type fixtureLeague struct {
	ID     int    `json:"id"`
	Season int    `json:"season"`
	Round  string `json:"round"`
}
type fixtureTeams struct {
	Home struct {
		ID int `json:"id"`
	} `json:"home"`
	Away struct {
		ID int `json:"id"`
	} `json:"away"`
}
type fixtureGoals struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// ---------- API client ----------

func apiGet[T any](endpoint string, params map[string]string) (*apiResponse[T], error) {
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-apisports-key", apiKey)

	q := req.URL.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var result apiResponse[T]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	// Debug: log raw response when empty
	if len(result.Response) == 0 {
		log.Printf("  DEBUG: empty response from %s — raw: %s", req.URL.String(), string(body[:min(len(body), 500)]))
	}

	return &result, nil
}

// ---------- Import functions ----------

func importTeams(ctx context.Context) error {
	log.Println("=== Importing teams ===")

	resp, err := apiGet[teamResponse]("/teams", map[string]string{
		"league": fmt.Sprintf("%d", leagueID),
		"season": fmt.Sprintf("%d", season),
	})
	if err != nil {
		return fmt.Errorf("fetch teams: %w", err)
	}

	total := len(resp.Response)
	for i, t := range resp.Response {
		_, err := pool.Exec(ctx, `
			INSERT INTO teams (id, name, code, logo_url, country, founded, venue_name, venue_city)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				code = EXCLUDED.code,
				logo_url = EXCLUDED.logo_url,
				country = EXCLUDED.country,
				founded = EXCLUDED.founded,
				venue_name = EXCLUDED.venue_name,
				venue_city = EXCLUDED.venue_city,
				updated_at = NOW()`,
			t.Team.ID, t.Team.Name, t.Team.Code, t.Team.Logo,
			t.Team.Country, t.Team.Founded, t.Venue.Name, t.Venue.City,
		)
		if err != nil {
			log.Printf("  ERROR inserting team %s: %v", t.Team.Name, err)
			continue
		}
		log.Printf("  Imported %d/%d teams: %s", i+1, total, t.Team.Name)
	}

	log.Printf("=== Teams done: %d imported ===\n", total)
	return nil
}

func importPlayers(ctx context.Context) error {
	log.Println("=== Importing players ===")

	page := 1
	totalImported := 0

	for {
		resp, err := apiGet[playerResponse]("/players", map[string]string{
			"league": fmt.Sprintf("%d", leagueID),
			"season": fmt.Sprintf("%d", season),
			"page":   fmt.Sprintf("%d", page),
		})
		if err != nil {
			return fmt.Errorf("fetch players page %d: %w", page, err)
		}

		for _, p := range resp.Response {
			pos := mapPosition(p)
			teamID, teamName := extractTeam(p)

			firstName := p.Player.FirstName
			lastName := p.Player.LastName
			if firstName == "" {
				// Some players only have a "name" field
				parts := strings.SplitN(p.Player.Name, " ", 2)
				firstName = parts[0]
				if len(parts) > 1 {
					lastName = parts[1]
				} else {
					lastName = ""
				}
			}

			_, err := pool.Exec(ctx, `
				INSERT INTO players (api_football_id, first_name, last_name, position, team_name, team_id, photo_url, age, nationality, market_value, is_active)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE)
				ON CONFLICT (api_football_id) DO UPDATE SET
					first_name = EXCLUDED.first_name,
					last_name = EXCLUDED.last_name,
					position = EXCLUDED.position,
					team_name = EXCLUDED.team_name,
					team_id = EXCLUDED.team_id,
					photo_url = EXCLUDED.photo_url,
					age = EXCLUDED.age,
					nationality = EXCLUDED.nationality,
					updated_at = NOW()`,
				p.Player.ID, firstName, lastName, pos, teamName,
				nilIfZero(teamID), p.Player.Photo, nilIfZero(p.Player.Age),
				nilIfEmpty(p.Player.Nationality), defaultMarketValue(pos),
			)
			if err != nil {
				log.Printf("  ERROR inserting player %s %s: %v", firstName, lastName, err)
				continue
			}
			totalImported++
		}

		log.Printf("  Page %d/%d — %d players so far", resp.Paging.Current, resp.Paging.Total, totalImported)

		if resp.Paging.Current >= resp.Paging.Total {
			break
		}
		page++
		time.Sleep(1 * time.Second) // Rate limiting
	}

	log.Printf("=== Players done: %d imported ===\n", totalImported)
	return nil
}

func importFixtures(ctx context.Context) error {
	log.Println("=== Importing fixtures ===")

	resp, err := apiGet[fixtureResponse]("/fixtures", map[string]string{
		"league": fmt.Sprintf("%d", leagueID),
		"season": fmt.Sprintf("%d", season),
	})
	if err != nil {
		return fmt.Errorf("fetch fixtures: %w", err)
	}

	total := len(resp.Response)
	imported := 0

	for _, f := range resp.Response {
		matchDate, err := time.Parse(time.RFC3339, f.Fixture.Date)
		if err != nil {
			matchDate, err = time.Parse("2006-01-02T15:04:05-07:00", f.Fixture.Date)
			if err != nil {
				log.Printf("  ERROR parsing date for fixture %d: %v", f.Fixture.ID, err)
				continue
			}
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO fixtures (id, league_id_api, season, round, home_team_id, away_team_id, home_goals, away_goals, status, match_date, venue_name, venue_city)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				round = EXCLUDED.round,
				home_goals = EXCLUDED.home_goals,
				away_goals = EXCLUDED.away_goals,
				status = EXCLUDED.status,
				match_date = EXCLUDED.match_date,
				venue_name = EXCLUDED.venue_name,
				venue_city = EXCLUDED.venue_city,
				updated_at = NOW()`,
			f.Fixture.ID, f.League.ID, f.League.Season, f.League.Round,
			f.Teams.Home.ID, f.Teams.Away.ID,
			f.Goals.Home, f.Goals.Away,
			f.Fixture.Status.Short, matchDate,
			f.Fixture.Venue.Name, f.Fixture.Venue.City,
		)
		if err != nil {
			log.Printf("  ERROR inserting fixture %d: %v", f.Fixture.ID, err)
			continue
		}
		imported++
	}

	log.Printf("=== Fixtures done: %d/%d imported ===\n", imported, total)
	return nil
}

// ---------- Helpers ----------

func mapPosition(p playerResponse) string {
	if len(p.Statistics) > 0 && p.Statistics[0].Games.Position != "" {
		switch p.Statistics[0].Games.Position {
		case "Goalkeeper":
			return "GK"
		case "Defender":
			return "DEF"
		case "Midfielder":
			return "MID"
		case "Attacker":
			return "FWD"
		}
	}
	return "MID" // default
}

func extractTeam(p playerResponse) (int, string) {
	if len(p.Statistics) > 0 {
		return p.Statistics[0].Team.ID, p.Statistics[0].Team.Name
	}
	return 0, "Unknown"
}

func defaultMarketValue(pos string) int {
	switch pos {
	case "GK":
		return 1500000
	case "DEF":
		return 2000000
	case "MID":
		return 2500000
	case "FWD":
		return 3000000
	default:
		return 1000000
	}
}

func nilIfZero(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// ---------- Main ----------

func main() {
	_ = godotenv.Load()

	apiKey = os.Getenv("API_FOOTBALL_KEY")
	if apiKey == "" {
		log.Fatal("API_FOOTBALL_KEY not set in .env")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set in .env")
	}

	ctx := context.Background()

	var err error
	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Connected to database")

	// Import in order: teams → players → fixtures
	if err := importTeams(ctx); err != nil {
		log.Fatalf("Teams import failed: %v", err)
	}

	if err := importPlayers(ctx); err != nil {
		log.Fatalf("Players import failed: %v", err)
	}

	if err := importFixtures(ctx); err != nil {
		log.Fatalf("Fixtures import failed: %v", err)
	}

	log.Println("✅ All data imported successfully!")
}
