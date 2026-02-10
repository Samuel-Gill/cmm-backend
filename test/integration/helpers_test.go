package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"matchmaking-service/auth/password"
	authrepository "matchmaking-service/auth/repository"
	authservice "matchmaking-service/auth/service"
	"matchmaking-service/auth/token"
	matchesrepository "matchmaking-service/matches/repository"
	matchesservice "matchmaking-service/matches/service"
	notificationsrepository "matchmaking-service/notifications/repository"
	notificationsservice "matchmaking-service/notifications/service"
	profilesrepository "matchmaking-service/profiles/repository"
	profilesservice "matchmaking-service/profiles/service"
	"matchmaking-service/routing"
)

var (
	migrationsOnce sync.Once
	migrationsErr  error
)

type integrationEnv struct {
	db  *sql.DB
	app http.Handler
}

type authTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func setupIntegration(t *testing.T) integrationEnv {
	t.Helper()
	db := openTestDB(t)
	applyMigrations(t, db)
	truncateAll(t, db)

	hasher := password.BCryptHasher{Cost: 4}
	authSvc := authservice.New(
		authrepository.NewPostgresStore(db),
		hasher,
		token.NewManager("integration-test-secret"),
		authservice.Config{AccessTokenTTL: time.Hour, RefreshTokenTTL: 24 * time.Hour, ResetPasswordTokenTTL: time.Hour},
	)
	profilesSvc := profilesservice.New(profilesrepository.NewPostgresStore(db))
	matchesSvc := matchesservice.New(matchesrepository.NewPostgresStore(db))
	notificationsSvc := notificationsservice.New(notificationsrepository.NewPostgresStore(db))

	return integrationEnv{db: db, app: routing.NewRouter(authSvc, profilesSvc, matchesSvc, notificationsSvc)}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/matchmaking_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	migrationsOnce.Do(func() {
		_, currentFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		migrationFiles := []string{
			"storage/migrations/V1__create_extensions.sql",
			"storage/migrations/V2__create_matchmaking_tables.sql",
			"storage/migrations/V3__create_indexes.sql",
			"storage/migrations/V4__create_password_reset_table.sql",
			"storage/migrations/V5__extend_profiles_for_management.sql",
			"storage/migrations/V6__matchmaking_filter_indexes.sql",
			"storage/migrations/V7__create_notifications.sql",
		}
		for _, rel := range migrationFiles {
			content, err := os.ReadFile(filepath.Join(repoRoot, rel))
			if err != nil {
				migrationsErr = fmt.Errorf("read %s: %w", rel, err)
				return
			}
			if _, err := db.Exec(string(content)); err != nil {
				migrationsErr = fmt.Errorf("exec %s: %w", rel, err)
				return
			}
		}
		// Test-only compatibility shim for profile INSERT statements used by handlers.
		_, migrationsErr = db.Exec(`
			ALTER TABLE matchmaking_profiles ALTER COLUMN display_name DROP NOT NULL;
			ALTER TABLE matchmaking_profiles ALTER COLUMN gender DROP NOT NULL;
			ALTER TABLE matchmaking_profiles ALTER COLUMN birth_date DROP NOT NULL;
		`)
	})
	if migrationsErr != nil {
		t.Fatalf("migrations failed: %v", migrationsErr)
	}
}

func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE TABLE
			notifications,
			matches,
			profile_likes,
			profile_filter_exclusions,
			search_filters,
			profile_interests,
			matchmaking_profiles,
			auth_password_resets,
			auth_refresh_tokens,
			auth_users
		RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func requestJSON(t *testing.T, app http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res := httptest.NewRecorder()
	app.ServeHTTP(res, req)
	return res
}

func decodeJSON[T any](t *testing.T, res *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, res.Body.String())
	}
	return out
}

func signupAndLogin(t *testing.T, app http.Handler, email, password string) (string, authTokens) {
	t.Helper()
	signupRes := requestJSON(t, app, http.MethodPost, "/auth/signup", map[string]string{"email": email, "password": password}, "")
	if signupRes.Code != http.StatusOK {
		t.Fatalf("signup failed: status=%d body=%s", signupRes.Code, signupRes.Body.String())
	}
	var signup map[string]string
	signup = decodeJSON[map[string]string](t, signupRes)

	loginRes := requestJSON(t, app, http.MethodPost, "/auth/login", map[string]string{"email": email, "password": password}, "")
	if loginRes.Code != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%s", loginRes.Code, loginRes.Body.String())
	}
	return signup["id"], decodeJSON[authTokens](t, loginRes)
}

func createProfile(t *testing.T, app http.Handler, accessToken, userID string, age, income int, location string, privacy map[string]bool) {
	t.Helper()
	payload := map[string]any{
		"user_id":           userID,
		"age":               age,
		"profession":        "Engineer",
		"education":         "Masters",
		"income":            income,
		"residency_status":  "citizen",
		"location":          location,
		"marital_status":    "single",
		"description":       "profile description",
		"hide_contact_info": privacy["hide_contact_info"],
		"hide_address":      privacy["hide_address"],
		"hide_income":       privacy["hide_income"],
		"hide_visa_status":  privacy["hide_visa_status"],
	}
	res := requestJSON(t, app, http.MethodPost, "/profiles/", payload, accessToken)
	if res.Code != http.StatusCreated {
		t.Fatalf("create profile failed: status=%d body=%s", res.Code, res.Body.String())
	}
}
