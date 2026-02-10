package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"matchmaking-service/auth/model"
	"matchmaking-service/auth/password"
	"matchmaking-service/auth/repository"
	"matchmaking-service/auth/service"
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

type testStore struct {
	users     map[string]model.User
	usersByID map[string]model.User
	refresh   map[string]model.RefreshToken
	resets    map[string]model.PasswordReset
	counter   int
}

func newTestStore() *testStore {
	return &testStore{users: map[string]model.User{}, usersByID: map[string]model.User{}, refresh: map[string]model.RefreshToken{}, resets: map[string]model.PasswordReset{}}
}

func (s *testStore) CreateUser(_ context.Context, email, passwordHash string) (model.User, error) {
	s.counter++
	u := model.User{ID: fmt.Sprintf("user-%d", s.counter), Email: email, PasswordHash: passwordHash, AccountState: "active", CreatedAt: time.Now()}
	s.users[email] = u
	s.usersByID[u.ID] = u
	return u, nil
}
func (s *testStore) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	u, ok := s.users[email]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}
func (s *testStore) GetUserByID(_ context.Context, id string) (model.User, error) {
	u, ok := s.usersByID[id]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}
func (s *testStore) UpdatePassword(_ context.Context, userID, passwordHash string) error {
	u := s.usersByID[userID]
	u.PasswordHash = passwordHash
	s.usersByID[userID] = u
	s.users[u.Email] = u
	return nil
}
func (s *testStore) SaveRefreshToken(_ context.Context, t model.RefreshToken) error {
	s.refresh[t.TokenHash] = t
	return nil
}
func (s *testStore) RevokeRefreshToken(_ context.Context, tokenHash string) error {
	t := s.refresh[tokenHash]
	now := time.Now()
	t.RevokedAt = &now
	s.refresh[tokenHash] = t
	return nil
}
func (s *testStore) GetActiveRefreshToken(_ context.Context, tokenHash string) (model.RefreshToken, error) {
	t, ok := s.refresh[tokenHash]
	if !ok {
		return model.RefreshToken{}, repository.ErrNotFound
	}
	return t, nil
}
func (s *testStore) SavePasswordReset(_ context.Context, reset model.PasswordReset) error {
	s.resets[reset.TokenHash] = reset
	return nil
}
func (s *testStore) ConsumePasswordReset(_ context.Context, tokenHash string) (model.PasswordReset, error) {
	r, ok := s.resets[tokenHash]
	if !ok {
		return model.PasswordReset{}, repository.ErrNotFound
	}
	now := time.Now()
	r.UsedAt = &now
	s.resets[tokenHash] = r
	return r, nil
}

type noopProfileStore struct{}

func (noopProfileStore) Create(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	return p, nil
}
func (noopProfileStore) Update(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	return p, nil
}
func (noopProfileStore) GetByID(_ context.Context, _ string) (profilesmodel.Profile, error) {
	return profilesmodel.Profile{}, profilesrepo.ErrNotFound
}
func (noopProfileStore) List(_ context.Context, _, _ int) ([]profilesmodel.Profile, error) {
	return nil, nil
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

func setup(t *testing.T) (*service.Service, http.Handler) {
	t.Helper()
	store := newTestStore()
	hasher := password.BCryptHasher{Cost: 4}
	svc := service.New(store, hasher, token.NewManager("test-secret"), service.Config{
		AccessTokenTTL:        time.Hour,
		RefreshTokenTTL:       time.Hour,
		ResetPasswordTokenTTL: time.Hour,
	})
	profilesSvc := profilesservice.New(noopProfileStore{})
	matchesSvc := matchesservice.New(noopMatchesStore{})
	notificationsSvc := notificationsservice.New(noopNotificationsStore{})
	return svc, routing.NewRouter(svc, profilesSvc, matchesSvc, notificationsSvc)
}

func TestSignupEndpoint(t *testing.T) {
	_, r := setup(t)
	body := []byte(`{"email":"a@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}

func TestLoginEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	body := []byte(`{"email":"a@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}

func TestLogoutEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := svc.Login(context.Background(), "a@example.com", "password123")
	payload, _ := json.Marshal(map[string]string{"refresh_token": toks.RefreshToken})
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(payload))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", res.Code)
	}
}

func TestRequestResetPasswordEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/request", bytes.NewReader([]byte(`{"email":"a@example.com"}`)))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}

func TestResetPasswordEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	resetToken, _ := svc.RequestPasswordReset(context.Background(), "a@example.com")
	payload, _ := json.Marshal(map[string]string{"token": resetToken, "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/password/reset/confirm", bytes.NewReader(payload))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", res.Code)
	}
}

func TestChangePasswordEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := svc.Login(context.Background(), "a@example.com", "password123")
	payload, _ := json.Marshal(map[string]string{"old_password": "password123", "new_password": "newpassword123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/password/change", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+toks.AccessToken)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", res.Code)
	}
}

func TestProtectedMeEndpoint(t *testing.T) {
	svc, r := setup(t)
	_, _ = svc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := svc.Login(context.Background(), "a@example.com", "password123")
	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+toks.AccessToken)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}
