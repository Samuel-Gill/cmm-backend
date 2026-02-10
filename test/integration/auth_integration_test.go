package integration_test

import (
	"database/sql"
	"net/http"
	"testing"

	authrepository "matchmaking-service/auth/repository"
	"matchmaking-service/auth/token"
)

func TestAuthSignupAndLogin(t *testing.T) {
	env := setupIntegration(t)

	cases := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{name: "valid", email: "signup-login@example.com", password: "Password123", wantStatus: http.StatusOK},
		{name: "invalid password", email: "short-pass@example.com", password: "short", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := requestJSON(t, env.app, http.MethodPost, "/auth/signup", map[string]string{"email": tc.email, "password": tc.password}, "")
			if res.Code != tc.wantStatus {
				t.Fatalf("signup status=%d body=%s", res.Code, res.Body.String())
			}
			if tc.wantStatus != http.StatusOK {
				return
			}

			loginRes := requestJSON(t, env.app, http.MethodPost, "/auth/login", map[string]string{"email": tc.email, "password": tc.password}, "")
			if loginRes.Code != http.StatusOK {
				t.Fatalf("login status=%d body=%s", loginRes.Code, loginRes.Body.String())
			}
			tokens := decodeJSON[authTokens](t, loginRes)
			if tokens.AccessToken == "" || tokens.RefreshToken == "" {
				t.Fatalf("expected both access and refresh token")
			}
		})
	}
}

func TestAuthRefreshTokenLifecycle(t *testing.T) {
	env := setupIntegration(t)
	_, tokens := signupAndLogin(t, env.app, "refresh-flow@example.com", "Password123")

	_, refreshHash, err := token.NewOpaqueTokenFromPlain(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("hash refresh token: %v", err)
	}

	store := authrepository.NewPostgresStore(env.db)
	if _, err := store.GetActiveRefreshToken(t.Context(), refreshHash); err != nil {
		t.Fatalf("expected active refresh token: %v", err)
	}

	logoutRes := requestJSON(t, env.app, http.MethodPost, "/auth/logout", map[string]string{"refresh_token": tokens.RefreshToken}, "")
	if logoutRes.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", logoutRes.Code, logoutRes.Body.String())
	}

	_, err = store.GetActiveRefreshToken(t.Context(), refreshHash)
	if err == nil {
		t.Fatalf("expected refresh token to be revoked")
	}
}

func TestAuthChangePassword(t *testing.T) {
	env := setupIntegration(t)
	_, tokens := signupAndLogin(t, env.app, "change-pass@example.com", "Password123")

	changeRes := requestJSON(t, env.app, http.MethodPost, "/auth/password/change", map[string]string{
		"old_password": "Password123",
		"new_password": "NewPassword123",
	}, tokens.AccessToken)
	if changeRes.Code != http.StatusNoContent {
		t.Fatalf("change password status=%d body=%s", changeRes.Code, changeRes.Body.String())
	}

	oldLogin := requestJSON(t, env.app, http.MethodPost, "/auth/login", map[string]string{"email": "change-pass@example.com", "password": "Password123"}, "")
	if oldLogin.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password rejected, status=%d", oldLogin.Code)
	}

	newLogin := requestJSON(t, env.app, http.MethodPost, "/auth/login", map[string]string{"email": "change-pass@example.com", "password": "NewPassword123"}, "")
	if newLogin.Code != http.StatusOK {
		t.Fatalf("expected new password accepted, status=%d body=%s", newLogin.Code, newLogin.Body.String())
	}
}

func TestAuthResetPasswordTokenFlow(t *testing.T) {
	env := setupIntegration(t)
	_, _ = signupAndLogin(t, env.app, "reset-flow@example.com", "Password123")

	requestRes := requestJSON(t, env.app, http.MethodPost, "/auth/password/reset/request", map[string]string{"email": "reset-flow@example.com"}, "")
	if requestRes.Code != http.StatusOK {
		t.Fatalf("reset request status=%d body=%s", requestRes.Code, requestRes.Body.String())
	}
	resetResponse := decodeJSON[map[string]string](t, requestRes)
	if resetResponse["reset_token"] == "" {
		t.Fatalf("expected reset_token")
	}

	confirmRes := requestJSON(t, env.app, http.MethodPost, "/auth/password/reset/confirm", map[string]string{
		"token":        resetResponse["reset_token"],
		"new_password": "ResetPassword123",
	}, "")
	if confirmRes.Code != http.StatusNoContent {
		t.Fatalf("reset confirm status=%d body=%s", confirmRes.Code, confirmRes.Body.String())
	}

	secondConfirm := requestJSON(t, env.app, http.MethodPost, "/auth/password/reset/confirm", map[string]string{
		"token":        resetResponse["reset_token"],
		"new_password": "AnotherPass123",
	}, "")
	if secondConfirm.Code != http.StatusBadRequest {
		t.Fatalf("expected single-use reset token, status=%d", secondConfirm.Code)
	}

	newLogin := requestJSON(t, env.app, http.MethodPost, "/auth/login", map[string]string{"email": "reset-flow@example.com", "password": "ResetPassword123"}, "")
	if newLogin.Code != http.StatusOK {
		t.Fatalf("expected login with reset password, status=%d body=%s", newLogin.Code, newLogin.Body.String())
	}
}

func TestAuthMeProtected(t *testing.T) {
	env := setupIntegration(t)
	userID, tokens := signupAndLogin(t, env.app, "me-route@example.com", "Password123")

	res := requestJSON(t, env.app, http.MethodGet, "/auth/me", nil, tokens.AccessToken)
	if res.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", res.Code, res.Body.String())
	}
	payload := decodeJSON[map[string]string](t, res)
	if payload["user_id"] != userID {
		t.Fatalf("user_id mismatch: got=%s want=%s", payload["user_id"], userID)
	}
}

func TestAuthRefreshTokenPersistedInDB(t *testing.T) {
	env := setupIntegration(t)
	_, tokens := signupAndLogin(t, env.app, "refresh-db@example.com", "Password123")
	_, refreshHash, err := token.NewOpaqueTokenFromPlain(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("hash refresh token: %v", err)
	}
	var exists bool
	err = env.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM auth_refresh_tokens WHERE token_hash = $1)`, refreshHash).Scan(&exists)
	if err != nil {
		t.Fatalf("query refresh token: %v", err)
	}
	if !exists {
		t.Fatalf("expected refresh token row")
	}
	var revokedAt sql.NullTime
	err = env.db.QueryRow(`SELECT revoked_at FROM auth_refresh_tokens WHERE token_hash = $1`, refreshHash).Scan(&revokedAt)
	if err != nil {
		t.Fatalf("query revoked_at: %v", err)
	}
	if revokedAt.Valid {
		t.Fatalf("expected fresh token to be not revoked")
	}
}
