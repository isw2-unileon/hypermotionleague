package apifootball

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testKey = "test-key-123"

// testClient spins up an httptest server with the given handler and returns a
// Client pointed at it (white-box: we set the unexported baseURL directly).
func testClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient(testKey)
	c.baseURL = srv.URL
	return c
}

// loadFixture reads a testdata file in the test goroutine (so a failure can use
// Fatalf safely, unlike inside the server handler goroutine).
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func TestGetTeams(t *testing.T) {
	fixture := loadFixture(t, "teams.json")
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/teams" {
			t.Errorf("path = %s, want /teams", r.URL.Path)
		}
		if got := r.Header.Get("x-apisports-key"); got != testKey {
			t.Errorf("auth header = %q, want %q", got, testKey)
		}
		q := r.URL.Query()
		if q.Get("league") != "141" || q.Get("season") != "2025" {
			t.Errorf("query = %v, want league=141 season=2025", q.Encode())
		}
		_, _ = w.Write(fixture)
	})

	teams, err := c.GetTeams(context.Background(), 141, 2025)
	if err != nil {
		t.Fatalf("GetTeams: %v", err)
	}
	if len(teams) != 3 {
		t.Fatalf("len(teams) = %d, want 3", len(teams))
	}

	t0 := teams[0]
	if t0.ID != 9001 || t0.Name != "Atlético Fictional" || t0.Code != "AFI" ||
		t0.Country != "Spain" || t0.Founded != 1925 ||
		t0.Logo != "https://media.api-sports.io/football/teams/9001.png" {
		t.Errorf("teams[0] = %+v, unexpected", t0)
	}
	if teams[1].Founded != 0 { // JSON null -> 0
		t.Errorf("teams[1].Founded = %d, want 0 (from null)", teams[1].Founded)
	}
	if teams[2].Code != "" { // JSON null -> ""
		t.Errorf("teams[2].Code = %q, want empty (from null)", teams[2].Code)
	}
}

func TestGetSquad(t *testing.T) {
	fixture := loadFixture(t, "squad.json")
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/players/squads" {
			t.Errorf("path = %s, want /players/squads", r.URL.Path)
		}
		if got := r.URL.Query().Get("team"); got != "9001" {
			t.Errorf("team query = %q, want 9001", got)
		}
		_, _ = w.Write(fixture)
	})

	players, err := c.GetSquad(context.Background(), 9001)
	if err != nil {
		t.Fatalf("GetSquad: %v", err)
	}
	if len(players) != 4 {
		t.Fatalf("len(players) = %d, want 4", len(players))
	}

	p0 := players[0]
	if p0.ID != 70001 || p0.Name != "Test Goalkeeper" || p0.Position != "Goalkeeper" ||
		p0.Number != 1 || p0.Age != 29 ||
		p0.Photo != "https://media.api-sports.io/football/players/70001.png" {
		t.Errorf("players[0] = %+v, unexpected", p0)
	}
	if players[1].Number != 0 { // JSON null -> 0
		t.Errorf("players[1].Number = %d, want 0 (from null)", players[1].Number)
	}
	if players[2].Age != 0 { // JSON null -> 0
		t.Errorf("players[2].Age = %d, want 0 (from null)", players[2].Age)
	}
	if players[3].Name != "Mononym" || players[3].Position != "Attacker" {
		t.Errorf("players[3] = %+v, unexpected", players[3])
	}
}

func TestGetTeamsAPIErrorObject(t *testing.T) {
	// errors as an object (e.g. plan/quota problems return this shape).
	const body = `{"get":"teams","parameters":{},"errors":{"requests":"You have reached the request limit"},"results":0,"response":[]}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	})

	_, err := c.GetTeams(context.Background(), 141, 2025)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "request limit") {
		t.Errorf("error = %v, want it to mention the request-limit message", err)
	}
}

func TestGetTeamsAPIErrorArray(t *testing.T) {
	// errors as an array of strings.
	const body = `{"get":"teams","parameters":{},"errors":["Bad request"],"results":0,"response":[]}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	})

	_, err := c.GetTeams(context.Background(), 141, 2025)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Bad request") {
		t.Errorf("error = %v, want it to mention 'Bad request'", err)
	}
}

func TestGetTeamsEmptyResponse(t *testing.T) {
	// No errors but no data -> loud failure (plan likely doesn't cover it).
	const body = `{"get":"teams","parameters":{},"errors":[],"results":0,"paging":{"current":1,"total":1},"response":[]}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	})

	_, err := c.GetTeams(context.Background(), 141, 2025)
	if err == nil {
		t.Fatal("expected loud error on empty response, got nil")
	}
	if !strings.Contains(err.Error(), "empty teams response") {
		t.Errorf("error = %v, want 'empty teams response'", err)
	}
}

func TestGetSquadEmptyResponse(t *testing.T) {
	// An empty squad block yields a nil slice and no error.
	const body = `{"get":"players/squads","parameters":{},"errors":[],"results":0,"response":[]}`
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, body)
	})

	players, err := c.GetSquad(context.Background(), 9001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if players != nil {
		t.Errorf("players = %v, want nil", players)
	}
}

func TestGetNon200Status(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "forbidden")
	})

	_, err := c.GetTeams(context.Background(), 141, 2025)
	if err == nil {
		t.Fatal("expected error on HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %v, want it to mention status 403", err)
	}
}

func TestAPIErrors(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"empty array", `[]`, false},
		{"empty object", `{}`, false},
		{"null", `null`, false},
		{"whitespace", `  `, false},
		{"object form", `{"token":"invalid api key"}`, true},
		{"array form", `["bad request","try again"]`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, ok := apiErrors(json.RawMessage(tc.raw))
			if ok != tc.wantErr {
				t.Errorf("apiErrors(%s) ok = %v, want %v (msg=%q)", tc.raw, ok, tc.wantErr, msg)
			}
		})
	}
}
