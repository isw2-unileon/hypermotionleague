package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

type mockPlayerRepo struct {
	// Lo que registramos en cada llamada a List, para comprobar que el
	// handler pasó bien los filtros al repo.
	gotPosition *models.PlayerPosition
	gotTeamName *string

	// Lo que el mock devolverá cuando se llame a List.
	returnPlayers []models.Player
	returnErr     error
}

func (m *mockPlayerRepo) List(_ context.Context, position *models.PlayerPosition, teamName *string) ([]models.Player, error) {
	m.gotPosition = position
	m.gotTeamName = teamName
	return m.returnPlayers, m.returnErr
}

// Los demás métodos no se usan en estos tests — devuelven valores nulos.
func (m *mockPlayerRepo) Create(_ context.Context, _ *models.Player) error { return nil }

func (m *mockPlayerRepo) GetByID(_ context.Context, _ int64) (*models.Player, error) {
	return nil, nil
}

func (m *mockPlayerRepo) GetAvailableForLeague(_ context.Context, _ int64) ([]models.Player, error) {
	return nil, nil
}
func (m *mockPlayerRepo) Update(_ context.Context, _ *models.Player) error             { return nil }
func (m *mockPlayerRepo) Delete(_ context.Context, _ int64) error                      { return nil }
func (m *mockPlayerRepo) UpsertByExternalID(_ context.Context, _ *models.Player) error { return nil }
func (m *mockPlayerRepo) UpsertPoints(_ context.Context, _ *models.PlayerPoints) error { return nil }
func (m *mockPlayerRepo) GetPoints(_ context.Context, _, _ int64) (*models.PlayerPoints, error) {
	return nil, nil
}

func (m *mockPlayerRepo) GetPointsByMatchday(_ context.Context, _ int64) ([]models.PlayerPoints, error) {
	return nil, nil
}

// Helper: monta un router de Gin con el handler bajo test y devuelve
// también el mock para poder inspeccionarlo después de la petición.
func newPlayerListRouter(t *testing.T, returnPlayers []models.Player) (*gin.Engine, *mockPlayerRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo := &mockPlayerRepo{returnPlayers: returnPlayers}
	// matchdays no se usa en List, así que pasamos nil sin problema.
	h := NewPlayerHandler(repo, nil)

	r := gin.New()
	r.GET("/api/v1/players", h.List)
	return r, repo
}

func TestPlayerHandler_List_PositionFilter(t *testing.T) {
	gk := models.PositionGK
	def := models.PositionDEF
	mid := models.PositionMID
	fwd := models.PositionFWD

	cases := []struct {
		name             string
		url              string
		expectedPosition *models.PlayerPosition
		expectedStatus   int
	}{
		{
			name:             "sin filtro pasa nil al repo",
			url:              "/api/v1/players",
			expectedPosition: nil,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "filtro GK pasa puntero a GK",
			url:              "/api/v1/players?position=GK",
			expectedPosition: &gk,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "filtro DEF pasa puntero a DEF",
			url:              "/api/v1/players?position=DEF",
			expectedPosition: &def,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "filtro MID pasa puntero a MID",
			url:              "/api/v1/players?position=MID",
			expectedPosition: &mid,
			expectedStatus:   http.StatusOK,
		},
		{
			name:             "filtro FWD pasa puntero a FWD",
			url:              "/api/v1/players?position=FWD",
			expectedPosition: &fwd,
			expectedStatus:   http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router, repo := newPlayerListRouter(t, []models.Player{})

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 1) El status code debe ser 200.
			if w.Code != tc.expectedStatus {
				t.Errorf("status esperado %d, obtenido %d (body=%s)", tc.expectedStatus, w.Code, w.Body.String())
			}

			// 2) El handler debe haber pasado el position correcto al repo.
			//    Comparamos puntero a puntero respetando el caso nil.
			if tc.expectedPosition == nil && repo.gotPosition != nil {
				t.Errorf("se esperaba nil position, pero el repo recibió %v", *repo.gotPosition)
			}
			if tc.expectedPosition != nil && repo.gotPosition == nil {
				t.Errorf("se esperaba position=%v, pero el repo recibió nil", *tc.expectedPosition)
			}
			if tc.expectedPosition != nil && repo.gotPosition != nil && *tc.expectedPosition != *repo.gotPosition {
				t.Errorf("position esperada %v, obtenida %v", *tc.expectedPosition, *repo.gotPosition)
			}
		})
	}
}

func TestPlayerHandler_List_TeamFilter(t *testing.T) {
	cases := []struct {
		name         string
		url          string
		expectedTeam *string
	}{
		{
			name:         "sin filtro team pasa nil",
			url:          "/api/v1/players",
			expectedTeam: nil,
		},
		{
			name:         "team con espacios codificados",
			url:          "/api/v1/players?team=Real+Madrid",
			expectedTeam: strPtr("Real Madrid"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router, repo := newPlayerListRouter(t, []models.Player{})

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("esperaba 200, obtuve %d", w.Code)
			}

			if tc.expectedTeam == nil && repo.gotTeamName != nil {
				t.Errorf("esperaba team nil, repo recibió %q", *repo.gotTeamName)
			}
			if tc.expectedTeam != nil && repo.gotTeamName == nil {
				t.Errorf("esperaba team=%q, repo recibió nil", *tc.expectedTeam)
			}
			if tc.expectedTeam != nil && repo.gotTeamName != nil && *tc.expectedTeam != *repo.gotTeamName {
				t.Errorf("team esperado %q, obtenido %q", *tc.expectedTeam, *repo.gotTeamName)
			}
		})
	}
}

func TestPlayerHandler_List_ResponseShape(t *testing.T) {
	players := []models.Player{
		{ID: 1, FirstName: "Iker", LastName: "Casillas", Position: models.PositionGK, TeamName: "Real Madrid"},
		{ID: 2, FirstName: "Sergio", LastName: "Ramos", Position: models.PositionDEF, TeamName: "Real Madrid"},
	}
	router, _ := newPlayerListRouter(t, players)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d (body=%s)", w.Code, w.Body.String())
	}

	var body struct {
		Players []models.Player `json:"players"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("no se pudo parsear el JSON de respuesta: %v", err)
	}

	if len(body.Players) != 2 {
		t.Errorf("esperaba 2 jugadores en la respuesta, obtuve %d", len(body.Players))
	}
	if body.Players[0].ID != 1 || body.Players[1].ID != 2 {
		t.Errorf("respuesta inesperada: %+v", body.Players)
	}
}

func strPtr(s string) *string {
	return &s
}
