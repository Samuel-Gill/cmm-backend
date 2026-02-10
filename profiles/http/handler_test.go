package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"matchmaking-service/auth/model"
	"matchmaking-service/auth/password"
	authrepo "matchmaking-service/auth/repository"
	authservice "matchmaking-service/auth/service"
	"matchmaking-service/auth/token"
	matchesmodel "matchmaking-service/matches/model"
	matchesservice "matchmaking-service/matches/service"
	notificationsmodel "matchmaking-service/notifications/model"
	notificationsservice "matchmaking-service/notifications/service"
	profilesmodel "matchmaking-service/profiles/model"
	profilesrepo "matchmaking-service/profiles/repository"
	profilesservice "matchmaking-service/profiles/service"
	"matchmaking-service/routing"
)

type authStore struct {
	users   map[string]model.User
	byID    map[string]model.User
	refresh map[string]model.RefreshToken
	resets  map[string]model.PasswordReset
}

func newAuthStore() *authStore {
	return &authStore{users: map[string]model.User{}, byID: map[string]model.User{}, refresh: map[string]model.RefreshToken{}, resets: map[string]model.PasswordReset{}}
}

func (s *authStore) CreateUser(_ context.Context, email, passwordHash string) (model.User, error) {
	u := model.User{ID: "u1", Email: email, PasswordHash: passwordHash, CreatedAt: time.Now()}
	s.users[email] = u
	s.byID[u.ID] = u
	return u, nil
}
func (s *authStore) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	u, ok := s.users[email]
	if !ok {
		return model.User{}, authrepo.ErrNotFound
	}
	return u, nil
}
func (s *authStore) GetUserByID(_ context.Context, id string) (model.User, error) {
	u, ok := s.byID[id]
	if !ok {
		return model.User{}, authrepo.ErrNotFound
	}
	return u, nil
}
func (s *authStore) UpdatePassword(_ context.Context, userID, passwordHash string) error {
	u := s.byID[userID]
	u.PasswordHash = passwordHash
	s.byID[userID] = u
	s.users[u.Email] = u
	return nil
}
func (s *authStore) SaveRefreshToken(_ context.Context, t model.RefreshToken) error {
	s.refresh[t.TokenHash] = t
	return nil
}
func (s *authStore) RevokeRefreshToken(_ context.Context, tokenHash string) error { return nil }
func (s *authStore) GetActiveRefreshToken(_ context.Context, tokenHash string) (model.RefreshToken, error) {
	return s.refresh[tokenHash], nil
}
func (s *authStore) SavePasswordReset(_ context.Context, reset model.PasswordReset) error {
	s.resets[reset.TokenHash] = reset
	return nil
}
func (s *authStore) ConsumePasswordReset(_ context.Context, tokenHash string) (model.PasswordReset, error) {
	return s.resets[tokenHash], nil
}

type profileStore struct {
	byID map[string]profilesmodel.Profile
}

func newProfileStore() *profileStore { return &profileStore{byID: map[string]profilesmodel.Profile{}} }
func (s *profileStore) Create(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	s.byID[p.UserID] = p
	return p, nil
}
func (s *profileStore) Update(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	ex, ok := s.byID[p.UserID]
	if !ok {
		return p, profilesrepo.ErrNotFound
	}
	p.CreatedAt = ex.CreatedAt
	p.UpdatedAt = time.Now()
	s.byID[p.UserID] = p
	return p, nil
}
func (s *profileStore) GetByID(_ context.Context, userID string) (profilesmodel.Profile, error) {
	p, ok := s.byID[userID]
	if !ok {
		return p, profilesrepo.ErrNotFound
	}
	return p, nil
}
func (s *profileStore) List(_ context.Context, _, _ int) ([]profilesmodel.Profile, error) {
	out := make([]profilesmodel.Profile, 0, len(s.byID))
	for _, v := range s.byID {
		out = append(out, v)
	}
	return out, nil
}

type noopMatchesStore struct{}

type noopNotificationsStore struct{}

func (noopNotificationsStore) ListUnread(_ context.Context, _ string, _ int) ([]notificationsmodel.Notification, error) {
	return nil, nil
}
func (noopNotificationsStore) MarkRead(_ context.Context, _, _ string) error { return nil }

func (noopMatchesStore) Browse(_ context.Context, _ string, _ matchesmodel.BrowseFilter) ([]matchesmodel.BrowseProfile, error) {
	return nil, nil
}
func (noopMatchesStore) LikeProfile(_ context.Context, _, _ string) (bool, error)    { return false, nil }
func (noopMatchesStore) LikedBy(_ context.Context, _ string) ([]string, error)       { return nil, nil }
func (noopMatchesStore) Liked(_ context.Context, _ string) ([]string, error)         { return nil, nil }
func (noopMatchesStore) MutualMatches(_ context.Context, _ string) ([]string, error) { return nil, nil }

func setupRouter(t *testing.T) (http.Handler, string) {
	t.Helper()
	authSvc := authservice.New(newAuthStore(), password.BCryptHasher{Cost: 4}, token.NewManager("secret"), authservice.Config{AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour, ResetPasswordTokenTTL: time.Hour})
	_, _ = authSvc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := authSvc.Login(context.Background(), "a@example.com", "password123")
	profilesSvc := profilesservice.New(newProfileStore())
	matchesSvc := matchesservice.New(noopMatchesStore{})
	notificationsSvc := notificationsservice.New(noopNotificationsStore{})
	return routing.NewRouter(authSvc, profilesSvc, matchesSvc, notificationsSvc), toks.AccessToken
}

func TestCreateUpdateGetListProfile(t *testing.T) {
	r, access := setupRouter(t)
	createPayload := map[string]any{
		"user_id": "u1", "age": 29, "profession": "Engineer", "education": "Masters", "income": 120000,
		"residency_status": "Citizen", "location": "Berlin", "marital_status": "Single", "description": "About me",
		"hide_contact_info": true, "hide_address": false, "hide_income": true, "hide_visa_status": false,
	}
	b, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost, "/profiles/", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+access)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create expected 201 got %d", res.Code)
	}

	update := createPayload
	update["profession"] = "Senior Engineer"
	ub, _ := json.Marshal(update)
	ureq := httptest.NewRequest(http.MethodPut, "/profiles/u1", bytes.NewReader(ub))
	ureq.Header.Set("Authorization", "Bearer "+access)
	ures := httptest.NewRecorder()
	r.ServeHTTP(ures, ureq)
	if ures.Code != http.StatusOK {
		t.Fatalf("update expected 200 got %d", ures.Code)
	}

	greq := httptest.NewRequest(http.MethodGet, "/profiles/u1", nil)
	gres := httptest.NewRecorder()
	r.ServeHTTP(gres, greq)
	if gres.Code != http.StatusOK {
		t.Fatalf("get expected 200 got %d", gres.Code)
	}

	lreq := httptest.NewRequest(http.MethodGet, "/profiles/?page=1&size=10", nil)
	lres := httptest.NewRecorder()
	r.ServeHTTP(lres, lreq)
	if lres.Code != http.StatusOK {
		t.Fatalf("list expected 200 got %d", lres.Code)
	}
}

func TestCannotModifyOtherUsersProfile(t *testing.T) {
	r, access := setupRouter(t)
	createPayload := map[string]any{
		"user_id": "other", "age": 29, "profession": "Engineer", "education": "Masters", "income": 120000,
		"residency_status": "Citizen", "location": "Berlin", "marital_status": "Single", "description": "About me",
	}
	b, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost, "/profiles/", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+access)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", res.Code)
	}
}
