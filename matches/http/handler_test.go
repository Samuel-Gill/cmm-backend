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

type matchesStore struct{}

type noopNotificationsStore struct{}

func (noopNotificationsStore) ListUnread(_ context.Context, _ string, _ int) ([]notificationsmodel.Notification, error) {
	return nil, nil
}
func (noopNotificationsStore) MarkRead(_ context.Context, _, _ string) error { return nil }

func (matchesStore) Browse(_ context.Context, _ string, _ matchesmodel.BrowseFilter) ([]matchesmodel.BrowseProfile, error) {
	return []matchesmodel.BrowseProfile{{UserID: "u2", Age: 28, Gender: "female", Income: 90000, Location: "Berlin", ResidencyStatus: "Citizen"}}, nil
}
func (matchesStore) LikeProfile(_ context.Context, _, _ string) (bool, error) { return true, nil }
func (matchesStore) LikedBy(_ context.Context, _ string) ([]string, error) {
	return []string{"u2"}, nil
}
func (matchesStore) Liked(_ context.Context, _ string) ([]string, error) { return []string{"u3"}, nil }
func (matchesStore) MutualMatches(_ context.Context, _ string) ([]string, error) {
	return []string{"u2"}, nil
}

func setup(t *testing.T) (http.Handler, string) {
	t.Helper()
	authSvc := authservice.New(&authStore{users: map[string]model.User{}, byID: map[string]model.User{}, refresh: map[string]model.RefreshToken{}}, password.BCryptHasher{Cost: 4}, token.NewManager("secret"), authservice.Config{AccessTokenTTL: time.Hour, RefreshTokenTTL: time.Hour, ResetPasswordTokenTTL: time.Hour})
	_, _ = authSvc.Signup(context.Background(), "a@example.com", "password123")
	toks, _ := authSvc.Login(context.Background(), "a@example.com", "password123")
	profilesSvc := profilesservice.New(noopProfiles{})
	matchesSvc := matchesservice.New(matchesStore{})
	notificationsSvc := notificationsservice.New(noopNotificationsStore{})
	return routing.NewRouter(authSvc, profilesSvc, matchesSvc, notificationsSvc), toks.AccessToken
}

func TestBrowseLikeRelationships(t *testing.T) {
	r, access := setup(t)

	breq := httptest.NewRequest(http.MethodGet, "/matches/browse?min_age=20&max_age=35&gender=female&min_income=50000&max_income=120000&location=Berlin&residency_status=Citizen&page=1&size=10", nil)
	breq.Header.Set("Authorization", "Bearer "+access)
	bres := httptest.NewRecorder()
	r.ServeHTTP(bres, breq)
	if bres.Code != http.StatusOK {
		t.Fatalf("browse expected 200 got %d", bres.Code)
	}

	payload, _ := json.Marshal(map[string]string{"liked_user_id": "u2"})
	lreq := httptest.NewRequest(http.MethodPost, "/matches/like", bytes.NewReader(payload))
	lreq.Header.Set("Authorization", "Bearer "+access)
	lres := httptest.NewRecorder()
	r.ServeHTTP(lres, lreq)
	if lres.Code != http.StatusOK {
		t.Fatalf("like expected 200 got %d", lres.Code)
	}

	rreq := httptest.NewRequest(http.MethodGet, "/matches/relationships", nil)
	rreq.Header.Set("Authorization", "Bearer "+access)
	rres := httptest.NewRecorder()
	r.ServeHTTP(rres, rreq)
	if rres.Code != http.StatusOK {
		t.Fatalf("relationships expected 200 got %d", rres.Code)
	}
}
