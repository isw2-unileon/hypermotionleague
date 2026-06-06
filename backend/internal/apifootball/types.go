package apifootball

// DTOs mirror the API-Football v3 JSON shapes. Only the fields we persist (plus
// the IDs we key on) are modelled — this package deliberately does not capture
// the full API surface.

// Team is the inner "team" object returned by /teams and embedded in
// /players/squads responses.
type Team struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`    // e.g. "RMA"; may be null/empty
	Country string `json:"country"` // may be null/empty
	Founded int    `json:"founded"` // may be null -> 0
	Logo    string `json:"logo"`    // media.api-sports.io URL
}

// Venue is the stadium block attached to each /teams entry. We do not persist
// it; it is modelled only so TeamWrapper decodes cleanly.
type Venue struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

// TeamWrapper is one element of the /teams response array: a team plus its venue.
type TeamWrapper struct {
	Team  Team  `json:"team"`
	Venue Venue `json:"venue"`
}

// SquadPlayer is one player entry in a /players/squads response.
//
// Note: /players/squads returns "name" as a single combined string
// (e.g. "Borja Iglesias"), not split into first/last name.
type SquadPlayer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`      // may be null -> 0
	Number   int    `json:"number"`   // jersey number; may be null -> 0
	Position string `json:"position"` // Goalkeeper | Defender | Midfielder | Attacker
	Photo    string `json:"photo"`    // media.api-sports.io URL
}

// SquadResponse is the single element of the /players/squads response array:
// the team plus its current squad.
type SquadResponse struct {
	Team    Team          `json:"team"`
	Players []SquadPlayer `json:"players"`
}

// PlayerMatchStatsDTO is one player's statistics for a single fixture, mapped
// from /fixtures/players. It carries ONLY the baremo's input fields (plus the
// API-Football IDs needed to match the row to a player loaded in our DB by
// external_id) — the dozens of other stat fields API-Football returns are
// ignored.
//
// The fields line up 1:1 with scoring.PlayerMatchStats EXCEPT OwnGoals, which
// /fixtures/players does not expose (own goals live in /fixtures/events). Task
// 1.3 should leave PlayerMatchStats.OwnGoals at 0 when building from this DTO.
//
// API-Football v3 omits a stat field entirely when it did not happen (there is
// no `cards.red: 0`, the key is simply absent), so every numeric field here is
// coalesced from a nullable wire value to 0 — see the decoding in client.go.
type PlayerMatchStatsDTO struct {
	PlayerExternalID int // API-Football player.id — upsert key for players.external_id
	TeamExternalID   int // API-Football team.id

	Minutes int
	Goals   int
	Assists int
	Yellow  int // yellow cards
	Red     int // red cards
	// GoalsConceded and CleanSheet both describe the same fact — the team's
	// goals-against in this fixture — so they share one source (the fixture
	// score), not the per-player goals.conceded stat, which API-Football only
	// reliably populates for goalkeepers. See client.go.
	GoalsConceded int
	CleanSheet    bool
	PensMissed    int
	PensSaved     int // GK
	Saves         int // GK
}

// fixtureListItem is one element of the /fixtures response. Only the fields
// FetchFixturePlayerStats needs are modelled: the fixture ID (to fetch its
// player stats) and the final score per side (to compute each team's
// goals-against, which drives the clean-sheet derivation).
type fixtureListItem struct {
	Fixture struct {
		ID int `json:"id"`
	} `json:"fixture"`
	Teams struct {
		Home struct {
			ID int `json:"id"`
		} `json:"home"`
		Away struct {
			ID int `json:"id"`
		} `json:"away"`
	} `json:"teams"`
	Goals struct {
		Home *int `json:"home"` // null before kickoff
		Away *int `json:"away"`
	} `json:"goals"`
}

// fixturePlayersTeamBlock is one element of the /fixtures/players response: a
// team and its players' per-fixture statistics.
type fixturePlayersTeamBlock struct {
	Team struct {
		ID int `json:"id"`
	} `json:"team"`
	Players []struct {
		Player struct {
			ID int `json:"id"`
		} `json:"player"`
		Statistics []playerStatistics `json:"statistics"`
	} `json:"players"`
}

// playerStatistics is statistics[0] of a /fixtures/players player entry. Every
// numeric leaf is a *int so an absent key decodes to nil (then coalesced to 0)
// rather than being indistinguishable from a real 0.
type playerStatistics struct {
	Games struct {
		Minutes *int `json:"minutes"` // null when the player did not appear
	} `json:"games"`
	Goals struct {
		Total   *int `json:"total"`
		Assists *int `json:"assists"`
		Saves   *int `json:"saves"`
		// goals.conceded is intentionally not mapped: it is reliable only for
		// keepers. GoalsConceded comes from the fixture score instead.
	} `json:"goals"`
	Cards struct {
		Yellow *int `json:"yellow"`
		Red    *int `json:"red"`
	} `json:"cards"`
	Penalty struct {
		Missed *int `json:"missed"`
		Saved  *int `json:"saved"`
	} `json:"penalty"`
}
