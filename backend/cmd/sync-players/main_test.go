package main

import (
	"testing"

	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/apifootball"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

func TestMapPosition(t *testing.T) {
	cases := []struct {
		in   string
		want models.PlayerPosition
		ok   bool
	}{
		{"Goalkeeper", models.PositionGK, true},
		{"Defender", models.PositionDEF, true},
		{"Midfielder", models.PositionMID, true},
		{"Attacker", models.PositionFWD, true},
		{" Goalkeeper ", models.PositionGK, true}, // trimmed
		{"Coach", "", false},                      // unknown -> fallback
		{"", "", false},                           // empty -> fallback
		{"goalkeeper", "", false},                 // case-sensitive (API sends TitleCase)
	}
	for _, tc := range cases {
		got, ok := mapPosition(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("mapPosition(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSplitName(t *testing.T) {
	cases := []struct{ in, first, last string }{
		{"Borja Iglesias", "Borja", "Iglesias"},
		{"Jan Oblak Test", "Jan", "Oblak Test"}, // remainder kept whole
		{"Mononym", "", "Mononym"},              // single token -> last_name
		{"", "", ""},
		{"   ", "", ""},
		{"A  B", "A", "B"}, // collapses extra space in remainder
		{"  Leading Trailing  ", "Leading", "Trailing"},
	}
	for _, tc := range cases {
		f, l := splitName(tc.in)
		if f != tc.first || l != tc.last {
			t.Errorf("splitName(%q) = (%q, %q), want (%q, %q)", tc.in, f, l, tc.first, tc.last)
		}
	}
}

func TestNullableInt(t *testing.T) {
	if got := nullableInt(0); got != nil {
		t.Errorf("nullableInt(0) = %v, want nil", got)
	}
	if got := nullableInt(7); got == nil || *got != 7 {
		t.Errorf("nullableInt(7) = %v, want *7", got)
	}
}

func TestToClub(t *testing.T) {
	c := toClub(apifootball.Team{
		ID: 9001, Name: "Atlético Fictional", Code: "AFI",
		Country: "Spain", Founded: 1925, Logo: "logo.png",
	})
	if c.ExternalID != 9001 || c.Name != "Atlético Fictional" || c.Code != "AFI" ||
		c.Country != "Spain" || c.LogoURL != "logo.png" {
		t.Errorf("toClub = %+v, unexpected", c)
	}
	if c.Founded == nil || *c.Founded != 1925 {
		t.Errorf("toClub Founded = %v, want *1925", c.Founded)
	}

	// founded 0 (API null) -> nil so the column stores NULL
	if c2 := toClub(apifootball.Team{ID: 1, Founded: 0}); c2.Founded != nil {
		t.Errorf("toClub Founded for 0 = %v, want nil", c2.Founded)
	}
}

func TestToPlayer(t *testing.T) {
	pk := int64(42)

	// Zero number/age (API null) -> nil pointers (stored as NULL).
	p := toPlayer(
		apifootball.SquadPlayer{ID: 70002, Name: "Sample Defender", Age: 0, Number: 0, Photo: "p.png"},
		models.PositionDEF, "Atlético Fictional", &pk,
	)
	if p.ExternalID != 70002 || p.FirstName != "Sample" || p.LastName != "Defender" {
		t.Errorf("toPlayer name/id = %+v, unexpected", p)
	}
	if p.Position != models.PositionDEF || p.TeamName != "Atlético Fictional" || p.PhotoURL != "p.png" {
		t.Errorf("toPlayer = %+v, unexpected", p)
	}
	if !p.IsActive {
		t.Error("toPlayer IsActive = false, want true")
	}
	if p.TeamID == nil || *p.TeamID != 42 {
		t.Errorf("toPlayer TeamID = %v, want *42", p.TeamID)
	}
	if p.JerseyNumber != nil {
		t.Errorf("toPlayer JerseyNumber = %v, want nil for 0", p.JerseyNumber)
	}
	if p.Age != nil {
		t.Errorf("toPlayer Age = %v, want nil for 0", p.Age)
	}

	// Non-zero number/age set; mononym; nil team PK stays nil.
	p2 := toPlayer(
		apifootball.SquadPlayer{ID: 1, Name: "Mononym", Age: 22, Number: 9},
		models.PositionFWD, "T", nil,
	)
	if p2.JerseyNumber == nil || *p2.JerseyNumber != 9 {
		t.Errorf("toPlayer JerseyNumber = %v, want *9", p2.JerseyNumber)
	}
	if p2.Age == nil || *p2.Age != 22 {
		t.Errorf("toPlayer Age = %v, want *22", p2.Age)
	}
	if p2.FirstName != "" || p2.LastName != "Mononym" {
		t.Errorf("toPlayer mononym = (%q, %q), want (\"\", \"Mononym\")", p2.FirstName, p2.LastName)
	}
	if p2.TeamID != nil {
		t.Errorf("toPlayer TeamID = %v, want nil", p2.TeamID)
	}
}
