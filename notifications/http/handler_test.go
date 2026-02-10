package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"matchmaking-service/auth/model"
	"matchmaking-service/auth/password"
	authservice "matchmaking-service/auth/service"
	"matchmaking-service/auth/token"
	matchesmodel "matchmaking-service/matches/model"
	matchesservice "matchmaking-service/matches/service"
	notificationsmodel "matchmaking-service/notifications/model"
	notificationsservice "matchmaking-service/notifications/service"
	profilesmodel "matchmaking-service/profiles/model"
	profilesservice "matchmaking-service/profiles/service"
	"matchmaking-service/routing"
)

type authStore struct {
	users   map[string]model.User
	byID    map[string]model.User
	refresh map[string]model.RefreshToken
}

func (s *authStore) CreateUser(_ context.Context, email, passwordHash string) (model.User, error) {
	u := model.User{ID: "u1", Email: email, PasswordHash: passwordHash, CreatedAt: time.Now()}
	s.users[email] = u
	s.byID[u.ID] = u
	return u, nil
}
func (s *authStore) GetUserByEmail(_ context.Context, email string) (model.User, error) {
	return s.users[email], nil
}
func (s *authStore) GetUserByID(_ context.Context, id string) (model.User, error) {
	return s.byID[id], nil
}
func (s *authStore) UpdatePassword(_ context.Context, _, _ string) error { return nil }
func (s *authStore) SaveRefreshToken(_ context.Context, t model.RefreshToken) error {
	s.refresh[t.TokenHash] = t
	return nil
}
func (s *authStore) RevokeRefreshToken(_ context.Context, _ string) error { return nil }
func (s *authStore) GetActiveRefreshToken(_ context.Context, _ string) (model.RefreshToken, error) {
	return model.RefreshToken{}, nil
}
func (s *authStore) SavePasswordReset(_ context.Context, _ model.PasswordReset) error { return nil }
func (s *authStore) ConsumePasswordReset(_ context.Context, _ string) (model.PasswordReset, error) {
	return model.PasswordReset{}, nil
}

type noopProfiles struct{}

func (noopProfiles) Create(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	return p, nil
}
func (noopProfiles) Update(_ context.Context, p profilesmodel.Profile) (profilesmodel.Profile, error) {
	return p, nil
}
func (noopProfiles) GetByID(_ context.Context, _ string) (profilesmodel.Profile, error) {
	return profilesmodel.Profile{}, nil
}
func (noopProfiles) List(_ context.Context, _, _ int) ([]profilesmodel.Profile, error) {
	return nil, nil
}

type noopMatches struct{}

func (noopMatches) Browse(_ context.Context, _ string, _ matchesmodel.BrowseFilter) ([]matchesmodel.BrowseProfile, error) {
	return nil, nil
}
func (noopMatches) LikeProfile(_ context.Context, _, _ string) (bool, error)    { return false, nil }
func (noopMatches) LikedBy(_ context.Context, _ string) ([]string, error)       { return nil, nil }
func (noopMatches) Liked(_ context.Context, _ string) ([]string, error)         { return nil, nil }
func (noopMatches) MutualMatches(_ context.Context, _ string) ([]string, error) { return nil, nil }

type notificationsStore struct{}

func (notificationsStore) ListUnread(_ context.Context, _ string, _ int) ([]notificationsmodel.Notification, error) {
	return []notificationsmodel.Notification{{ID: "n1", UserID: "u1", Type: "liked", Title: "You got a new like"}}, nil
}
func (notificationsStore) MarkRead(_ context.Context, _, _ string) error { return nil }

func TestNotificationsEndpoints(t *testing.T) {
	authSvc := authservice.New(&authStore{users: map[string]model.User{}, byID: map[string]model.User{}, refresh: map[string]model.RefreshToken{}}, password.BCryptHasher{Cost: 4}, token.NewManager("secret"), authservice.Config{AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour, ResetPasswordTokenTTL: time.Hour})
	_, _ = authSvc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := authSvc.Login(context.Background(), "a@example.com", "password123")

	r := routing.NewRouter(
		authSvc,
		profilesservice.New(noopProfiles{}),
		matchesservice.New(noopMatches{}),
		notificationsservice.New(notificationsStore{}),
	)

	ureq := httptest.NewRequest(http.MethodGet, "/notifications/unread?limit=10", nil)
	ureq.Header.Set("Authorization", "Bearer "+toks.AccessToken)
	ures := httptest.NewRecorder()
	r.ServeHTTP(ures, ureq)
	if ures.Code != http.StatusOK {
		t.Fatalf("unread expected 200 got %d", ures.Code)
	}

	mreq := httptest.NewRequest(http.MethodPost, "/notifications/n1/read", nil)
	mreq.Header.Set("Authorization", "Bearer "+toks.AccessToken)
	mres := httptest.NewRecorder()
	r.ServeHTTP(mres, mreq)
	if mres.Code != http.StatusNoContent {
		t.Fatalf("mark read expected 204 got %d", mres.Code)
	}
}
