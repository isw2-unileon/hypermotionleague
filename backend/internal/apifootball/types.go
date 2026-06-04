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
