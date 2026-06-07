package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/proyect-scaffolding/backend/internal/models"
)

type mockLeagueRepoForCreate struct {
	// Programación de respuestas.
	createErr       error
	addMemberErr    error
	getByInviteCode *models.League
	getByInviteErr  error
	getMemberResult *models.LeagueMember
	getMemberErr    error
	countMembers    int
	countMembersErr error

	getByIDResult             *models.League
	getMembersWithUsersResult []models.LeagueMemberWithUser

	mu             sync.Mutex
	createdLeagues []*models.League
	addedMembers   []*models.LeagueMember
}

func (m *mockLeagueRepoForCreate) Create(_ context.Context, league *models.League) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	league.ID = 42
	m.createdLeagues = append(m.createdLeagues, league)
	return nil
}

func (m *mockLeagueRepoForCreate) AddMember(_ context.Context, member *models.LeagueMember) error {
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addedMembers = append(m.addedMembers, member)
	return nil
}

func (m *mockLeagueRepoForCreate) GetByInviteCode(_ context.Context, _ string) (*models.League, error) {
	return m.getByInviteCode, m.getByInviteErr
}

func (m *mockLeagueRepoForCreate) GetMember(_ context.Context, _, _ int64) (*models.LeagueMember, error) {
	return m.getMemberResult, m.getMemberErr
}

func (m *mockLeagueRepoForCreate) CountMembers(_ context.Context, _ int64) (int, error) {
	return m.countMembers, m.countMembersErr
}

func (m *mockLeagueRepoForCreate) GetByID(_ context.Context, _ int64) (*models.League, error) {
	return m.getByIDResult, nil
}

func (m *mockLeagueRepoForCreate) GetByUserID(_ context.Context, _ int64) ([]models.League, error) {
	return nil, nil
}
func (m *mockLeagueRepoForCreate) Update(_ context.Context, _ *models.League) error { return nil }
func (m *mockLeagueRepoForCreate) Delete(_ context.Context, _ int64) error          { return nil }
func (m *mockLeagueRepoForCreate) GetMembers(_ context.Context, _ int64) ([]models.LeagueMember, error) {
	return nil, nil
}

func (m *mockLeagueRepoForCreate) GetMembersWithUsers(_ context.Context, _ int64) ([]models.LeagueMemberWithUser, error) {
	return m.getMembersWithUsersResult, nil
}

func (m *mockLeagueRepoForCreate) UpdateMemberBudget(_ context.Context, _, _ int64, _ int) error {
	return nil
}
func (m *mockLeagueRepoForCreate) RemoveMember(_ context.Context, _, _ int64) error { return nil }

type mockTeamRepoForDraft struct {
	mu       sync.Mutex
	draftErr error
	calls    []draftCall
}

type draftCall struct {
	leagueID int64
	userID   int64
}

func (m *mockTeamRepoForDraft) DraftInitialSquad(_ context.Context, leagueID, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, draftCall{leagueID: leagueID, userID: userID})
	return m.draftErr
}

func (m *mockTeamRepoForDraft) callsCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func (m *mockTeamRepoForDraft) AddPlayer(_ context.Context, _ *models.TeamPlayer) error { return nil }
func (m *mockTeamRepoForDraft) RemovePlayer(_ context.Context, _, _, _ int64) error     { return nil }
func (m *mockTeamRepoForDraft) GetUserTeam(_ context.Context, _, _ int64) (*models.UserTeam, error) {
	return nil, nil
}

func (m *mockTeamRepoForDraft) GetPlayerOwner(_ context.Context, _, _ int64) (*models.TeamPlayer, error) {
	return nil, nil
}

func (m *mockTeamRepoForDraft) HasPlayer(_ context.Context, _, _, _ int64) (bool, error) {
	return false, nil
}

func (m *mockTeamRepoForDraft) TransferPlayer(_ context.Context, _, _, _, _ int64, _ int) error {
	return nil
}

func newLeagueRouter(t *testing.T, userID int64) (*gin.Engine, *mockLeagueRepoForCreate, *mockTeamRepoForDraft) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	leagueRepo := &mockLeagueRepoForCreate{}
	teamRepo := &mockTeamRepoForDraft{}
	// nil market-seeding collaborators: seedMarketIfOpen is a no-op here, so
	// these tests exercise only league creation + the initial squad draft.
	h := NewLeagueHandler(leagueRepo, teamRepo, nil, nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	r.POST("/api/v1/leagues", h.Create)
	r.POST("/api/v1/leagues/join", h.JoinLeague)

	return r, leagueRepo, teamRepo
}

func TestLeagueHandler_Create_InvokesDraft(t *testing.T) {
	const userID int64 = 7
	router, leagueRepo, teamRepo := newLeagueRouter(t, userID)

	body := models.CreateLeagueRequest{Name: "Mi Liga"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leagues", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuve %d (body=%s)", w.Code, w.Body.String())
	}

	if len(leagueRepo.createdLeagues) != 1 {
		t.Fatalf("esperaba 1 liga creada, hubo %d", len(leagueRepo.createdLeagues))
	}

	if len(leagueRepo.addedMembers) != 1 {
		t.Fatalf("esperaba 1 miembro añadido, hubo %d", len(leagueRepo.addedMembers))
	}
	if leagueRepo.addedMembers[0].Role != models.RoleOwner {
		t.Errorf("esperaba Role=owner, obtuve %v", leagueRepo.addedMembers[0].Role)
	}

	if teamRepo.callsCount() != 1 {
		t.Fatalf("esperaba 1 llamada a DraftInitialSquad, hubo %d", teamRepo.callsCount())
	}

	got := teamRepo.calls[0]
	if got.leagueID != 42 {
		t.Errorf("DraftInitialSquad recibió leagueID=%d, esperaba 42", got.leagueID)
	}
	if got.userID != userID {
		t.Errorf("DraftInitialSquad recibió userID=%d, esperaba %d", got.userID, userID)
	}
}

func TestLeagueHandler_Create_DraftFailure_StillReturnsCreated(t *testing.T) {
	router, _, teamRepo := newLeagueRouter(t, 7)
	teamRepo.draftErr = context.Canceled // cualquier error sirve

	body := models.CreateLeagueRequest{Name: "Liga sin jugadores"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/leagues", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("esperaba 201 incluso con draft fallido, obtuve %d", w.Code)
	}

	if teamRepo.callsCount() != 1 {
		t.Errorf("esperaba 1 intento de draft, hubo %d", teamRepo.callsCount())
	}
}

func TestLeagueHandler_Create_Unauthenticated(t *testing.T) {
	router, leagueRepo, teamRepo := newLeagueRouter(t, 0)

	body := models.CreateLeagueRequest{Name: "X"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/leagues", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperaba 401, obtuve %d", w.Code)
	}
	if len(leagueRepo.createdLeagues) != 0 {
		t.Errorf("no debería haberse creado liga, pero hay %d", len(leagueRepo.createdLeagues))
	}
	if teamRepo.callsCount() != 0 {
		t.Errorf("no debería haberse llamado al draft, pero hubo %d llamadas", teamRepo.callsCount())
	}
}

func TestLeagueHandler_JoinLeague_InvokesDraft(t *testing.T) {
	const userID int64 = 9
	router, leagueRepo, teamRepo := newLeagueRouter(t, userID)

	leagueRepo.getByInviteCode = &models.League{
		ID:            100,
		Name:          "Liga test",
		MaxMembers:    10,
		BudgetPerUser: 1000000,
	}
	leagueRepo.getMemberResult = nil
	leagueRepo.countMembers = 1

	body := models.JoinLeagueRequest{InviteCode: "ABCD1234"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/leagues/join", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d (body=%s)", w.Code, w.Body.String())
	}

	if len(leagueRepo.addedMembers) != 1 {
		t.Fatalf("esperaba 1 miembro añadido, hubo %d", len(leagueRepo.addedMembers))
	}
	if leagueRepo.addedMembers[0].Role != models.RoleMember {
		t.Errorf("esperaba Role=member, obtuve %v", leagueRepo.addedMembers[0].Role)
	}

	if teamRepo.callsCount() != 1 {
		t.Fatalf("esperaba 1 llamada a DraftInitialSquad, hubo %d", teamRepo.callsCount())
	}
	got := teamRepo.calls[0]
	if got.leagueID != 100 {
		t.Errorf("draft con leagueID=%d, esperaba 100", got.leagueID)
	}
	if got.userID != userID {
		t.Errorf("draft con userID=%d, esperaba %d", got.userID, userID)
	}
}

func TestLeagueHandler_JoinLeague_LeagueFull_NoDraft(t *testing.T) {
	router, leagueRepo, teamRepo := newLeagueRouter(t, 9)

	leagueRepo.getByInviteCode = &models.League{ID: 100, MaxMembers: 10}
	leagueRepo.countMembers = 10 // liga llena

	body := models.JoinLeagueRequest{InviteCode: "ABCD1234"}
	payload, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/leagues/join", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("esperaba 409, obtuve %d", w.Code)
	}
	if teamRepo.callsCount() != 0 {
		t.Errorf("no debería haber draft, pero hubo %d llamadas", teamRepo.callsCount())
	}
}

// newLeagueReadRouter wires the read endpoints (GetByID, GetMembers) with an
// authenticated caller, for the object-level authorization regression tests.
func newLeagueReadRouter(t *testing.T, userID int64) (*gin.Engine, *mockLeagueRepoForCreate) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	leagueRepo := &mockLeagueRepoForCreate{}
	h := NewLeagueHandler(leagueRepo, &mockTeamRepoForDraft{}, nil, nil, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	r.GET("/api/v1/leagues/:id", h.GetByID)
	r.GET("/api/v1/leagues/:id/members", h.GetMembers)
	return r, leagueRepo
}

// C1: a non-member must be rejected with 403, and the invite_code must NOT leak.
func TestLeagueHandler_GetByID_NonMember_Forbidden(t *testing.T) {
	router, repo := newLeagueReadRouter(t, 7)
	repo.getMemberResult = nil // caller is not a member of this league
	repo.getByIDResult = &models.League{ID: 1, Name: "Liga ajena", InviteCode: "s3cr3t"}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403 para no-miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "s3cr3t") {
		t.Fatalf("el invite_code no debe filtrarse a un no-miembro: %s", w.Body.String())
	}
}

func TestLeagueHandler_GetByID_Member_OK(t *testing.T) {
	router, repo := newLeagueReadRouter(t, 7)
	repo.getMemberResult = &models.LeagueMember{LeagueID: 1, UserID: 7}
	repo.getByIDResult = &models.League{ID: 1, Name: "Mi liga", InviteCode: "abc123"}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200 para miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}

// A1: the roster is only readable by members.
func TestLeagueHandler_GetMembers_NonMember_Forbidden(t *testing.T) {
	router, repo := newLeagueReadRouter(t, 7)
	repo.getMemberResult = nil

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/1/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403 para no-miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestLeagueHandler_GetMembers_Member_OK(t *testing.T) {
	router, repo := newLeagueReadRouter(t, 7)
	repo.getMemberResult = &models.LeagueMember{LeagueID: 1, UserID: 7}
	repo.getMembersWithUsersResult = []models.LeagueMemberWithUser{
		{LeagueMember: models.LeagueMember{LeagueID: 1, UserID: 7}, Username: "u7"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/leagues/1/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperaba 200 para miembro, obtuve %d (body=%s)", w.Code, w.Body.String())
	}
}
