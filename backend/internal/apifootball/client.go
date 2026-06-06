// Package apifootball is a thin HTTP client for the API-Football v3 API
// (https://v3.football.api-sports.io).
//
// It only speaks HTTP and returns typed DTOs — it contains no database logic.
// Politeness delays between calls are the caller's responsibility, not the
// client's, so the client stays a thin wrapper.
package apifootball

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the API-Football v3 base URL.
const DefaultBaseURL = "https://v3.football.api-sports.io"

// defaultTimeout bounds each HTTP call.
const defaultTimeout = 12 * time.Second

// defaultPoliteness is the pause inserted between the per-fixture calls
// FetchFixturePlayerStats makes, so resolving one round (~12 calls) does not
// hammer the API. It mirrors the delay cmd/sync-players already uses between
// squad calls.
const defaultPoliteness = 200 * time.Millisecond

// cleanSheetMinMinutes is the minutes-played threshold for a clean sheet to
// count, matching the scoring baremo.
const cleanSheetMinMinutes = 60

// Client talks to the API-Football v3 API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	politeness time.Duration // delay between FetchFixturePlayerStats per-fixture calls
}

// NewClient creates a Client authenticating with the given API-Football key.
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		politeness: defaultPoliteness,
	}
}

// envelope is the common API-Football response wrapper. The real data lives in
// Response; Errors is empty (`[]`) on success but may be an object on failure.
type envelope struct {
	Get      string          `json:"get"`
	Errors   json.RawMessage `json:"errors"`
	Results  int             `json:"results"`
	Response json.RawMessage `json:"response"`
}

// get issues a GET to path with the given query, validates the envelope, and
// returns the `response` array as raw messages for the caller to unmarshal.
func (c *Client) get(ctx context.Context, path string, query url.Values) ([]json.RawMessage, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("apifootball: build request %s: %w", path, err)
	}
	req.Header.Set("x-apisports-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apifootball: do request %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("apifootball: read body %s: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apifootball: %s returned status %d: %s", path, resp.StatusCode, snippet(body))
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("apifootball: decode envelope %s: %w", path, err)
	}

	// API-Football returns HTTP 200 even for auth/plan errors, surfacing them in
	// the errors field instead — so check it explicitly.
	if msg, ok := apiErrors(env.Errors); ok {
		return nil, fmt.Errorf("apifootball: %s error: %s", path, msg)
	}

	var items []json.RawMessage
	if len(env.Response) > 0 {
		if err := json.Unmarshal(env.Response, &items); err != nil {
			return nil, fmt.Errorf("apifootball: decode response array %s: %w", path, err)
		}
	}
	return items, nil
}

// GetTeams returns all teams in a league/season (a single, unpaginated call for
// a 22-team league).
func (c *Client) GetTeams(ctx context.Context, leagueID, season int) ([]Team, error) {
	q := url.Values{}
	q.Set("league", strconv.Itoa(leagueID))
	q.Set("season", strconv.Itoa(season))

	items, err := c.get(ctx, "/teams", q)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		// No errors but no data almost always means the plan does not cover this
		// league/season — fail loudly rather than silently load nothing.
		return nil, fmt.Errorf(
			"apifootball: empty teams response for league=%d season=%d "+
				"(does your API-Football plan cover this league/season?)", leagueID, season)
	}

	teams := make([]Team, 0, len(items))
	for i, raw := range items {
		var w TeamWrapper
		if err := json.Unmarshal(raw, &w); err != nil {
			return nil, fmt.Errorf("apifootball: decode team[%d]: %w", i, err)
		}
		teams = append(teams, w.Team)
	}
	return teams, nil
}

// GetSquad returns the current squad of one team (a single, unpaginated call).
// An empty response (no squad block) yields a nil slice rather than an error;
// the caller decides how to treat a team with zero players.
func (c *Client) GetSquad(ctx context.Context, teamID int) ([]SquadPlayer, error) {
	q := url.Values{}
	q.Set("team", strconv.Itoa(teamID))

	items, err := c.get(ctx, "/players/squads", q)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	// /players/squads returns a single element: {team, players[]}.
	var sr SquadResponse
	if err := json.Unmarshal(items[0], &sr); err != nil {
		return nil, fmt.Errorf("apifootball: decode squad for team=%d: %w", teamID, err)
	}
	return sr.Players, nil
}

// FetchFixturePlayerStats returns the per-player match statistics for every
// fixture in one round of the given league/season.
//
// API-Football has no "stats for a whole round" endpoint, so this makes two
// kinds of call:
//
//  1. GET /fixtures?league=&season=&round=Regular Season - N   (1 call)
//  2. GET /fixtures/players?fixture=ID                          (1 call per fixture)
//
// For Segunda (22 teams → 11 fixtures) that is ~12 calls per round. A short,
// cancellable politeness delay is inserted between the per-fixture calls (step
// 2); the caller does not need to pace it.
//
// Both GoalsConceded and CleanSheet are derived from the fixture result, not
// from per-player fields: a team's goals-against comes from step 1's score
// (home concedes the away goals, and vice-versa). The per-player goals.conceded
// stat is deliberately ignored because API-Football only reliably populates it
// for goalkeepers — using the team figure keeps both the "−1 per 2 conceded"
// penalty and the clean-sheet bonus correct for defenders too. A clean sheet is
// credited when the player was on the pitch for at least cleanSheetMinMinutes
// AND their team conceded zero.
func (c *Client) FetchFixturePlayerStats(ctx context.Context, league, season, round int) ([]PlayerMatchStatsDTO, error) {
	fixtures, err := c.getFixturesByRound(ctx, league, season, round)
	if err != nil {
		return nil, err
	}

	var out []PlayerMatchStatsDTO
	for i, fx := range fixtures {
		if i > 0 && c.politeness > 0 {
			// Politeness delay between the per-fixture calls (cancellable).
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.politeness):
			}
		}

		stats, err := c.getFixturePlayers(ctx, fx)
		if err != nil {
			return nil, fmt.Errorf("apifootball: fixture %d players: %w", fx.id, err)
		}
		out = append(out, stats...)
	}
	return out, nil
}

// roundFixture is the distilled fixture info FetchFixturePlayerStats needs: the
// fixture ID plus each team's goals-against in that fixture, keyed by
// API-Football team ID, used to derive clean sheets.
type roundFixture struct {
	id           int
	goalsAgainst map[int]int
}

// getFixturesByRound lists the fixtures of one round and distills each into a
// roundFixture (ID + per-team goals-against).
func (c *Client) getFixturesByRound(ctx context.Context, league, season, round int) ([]roundFixture, error) {
	q := url.Values{}
	q.Set("league", strconv.Itoa(league))
	q.Set("season", strconv.Itoa(season))
	q.Set("round", fmt.Sprintf("Regular Season - %d", round))

	items, err := c.get(ctx, "/fixtures", q)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		// No errors but no data: the plan likely does not cover this
		// league/season, or the round has not been scheduled. Fail loudly
		// rather than silently returning zero stats.
		return nil, fmt.Errorf(
			"apifootball: empty fixtures response for league=%d season=%d round=%d "+
				"(does your API-Football plan cover this league/season, and is the round scheduled?)",
			league, season, round)
	}

	fixtures := make([]roundFixture, 0, len(items))
	for i, raw := range items {
		var fi fixtureListItem
		if err := json.Unmarshal(raw, &fi); err != nil {
			return nil, fmt.Errorf("apifootball: decode fixture[%d]: %w", i, err)
		}
		fixtures = append(fixtures, roundFixture{
			id: fi.Fixture.ID,
			goalsAgainst: map[int]int{
				// A team concedes its opponent's goals.
				fi.Teams.Home.ID: intOr0(fi.Goals.Away),
				fi.Teams.Away.ID: intOr0(fi.Goals.Home),
			},
		})
	}
	return fixtures, nil
}

// getFixturePlayers fetches one fixture's per-player stats and maps each into a
// PlayerMatchStatsDTO, coalescing every absent numeric field to 0 and deriving
// clean sheet from the fixture's team goals-against.
func (c *Client) getFixturePlayers(ctx context.Context, fx roundFixture) ([]PlayerMatchStatsDTO, error) {
	q := url.Values{}
	q.Set("fixture", strconv.Itoa(fx.id))

	items, err := c.get(ctx, "/fixtures/players", q)
	if err != nil {
		return nil, err
	}

	var out []PlayerMatchStatsDTO
	for i, raw := range items {
		var block fixturePlayersTeamBlock
		if err := json.Unmarshal(raw, &block); err != nil {
			return nil, fmt.Errorf("apifootball: decode team block[%d] for fixture %d: %w", i, fx.id, err)
		}

		teamGA := fx.goalsAgainst[block.Team.ID]
		for _, p := range block.Players {
			if len(p.Statistics) == 0 {
				continue // no stats line for this player in this fixture
			}
			st := p.Statistics[0]
			minutes := intOr0(st.Games.Minutes)

			out = append(out, PlayerMatchStatsDTO{
				PlayerExternalID: p.Player.ID,
				TeamExternalID:   block.Team.ID,
				Minutes:          minutes,
				Goals:            intOr0(st.Goals.Total),
				Assists:          intOr0(st.Goals.Assists),
				Yellow:           intOr0(st.Cards.Yellow),
				Red:              intOr0(st.Cards.Red),
				GoalsConceded:    teamGA,
				CleanSheet:       minutes >= cleanSheetMinMinutes && teamGA == 0,
				PensMissed:       intOr0(st.Penalty.Missed),
				PensSaved:        intOr0(st.Penalty.Saved),
				Saves:            intOr0(st.Goals.Saves),
			})
		}
	}
	return out, nil
}

// intOr0 coalesces a nullable API-Football numeric field to 0. API-Football v3
// omits a stat entirely when it did not happen, so a nil pointer means "absent"
// and is treated as zero.
func intOr0(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// apiErrors reports whether the envelope's errors field signals a failure and,
// if so, a human-readable message. The field is `[]` on success but can be an
// object (e.g. {"token": "..."}) or array of strings on error.
func apiErrors(raw json.RawMessage) (string, bool) {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "[]" || s == "{}" || s == "null" {
		return "", false
	}

	// Object form: {"token": "msg", ...}
	var asMap map[string]string
	if err := json.Unmarshal(raw, &asMap); err == nil {
		if len(asMap) == 0 {
			return "", false
		}
		parts := make([]string, 0, len(asMap))
		for k, v := range asMap {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		}
		sort.Strings(parts) // deterministic message
		return strings.Join(parts, "; "), true
	}

	// Array form: ["msg1", "msg2"]
	var asArr []string
	if err := json.Unmarshal(raw, &asArr); err == nil {
		if len(asArr) == 0 {
			return "", false
		}
		return strings.Join(asArr, "; "), true
	}

	// Unknown but non-empty shape — surface the raw text.
	return s, true
}

// snippet trims a response body for inclusion in error messages.
func snippet(b []byte) string {
	const max = 200
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		return s[:max] + "…"
	}
	return s
}
