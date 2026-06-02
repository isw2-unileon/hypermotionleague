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

// Client talks to the API-Football v3 API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a Client authenticating with the given API-Football key.
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
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
