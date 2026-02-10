package integration_test

import (
	"net/http"
	"testing"
)

func TestMatchmakingLikeProfile(t *testing.T) {
	env := setupIntegration(t)
	actorID, actorTokens := signupAndLogin(t, env.app, "like-actor@example.com", "Password123")
	targetID, targetTokens := signupAndLogin(t, env.app, "like-target@example.com", "Password123")

	createProfile(t, env.app, actorTokens.AccessToken, actorID, 28, 70000, "NYC", map[string]bool{})
	createProfile(t, env.app, targetTokens.AccessToken, targetID, 29, 75000, "NYC", map[string]bool{})

	likeRes := requestJSON(t, env.app, http.MethodPost, "/matches/like", map[string]string{"liked_user_id": targetID}, actorTokens.AccessToken)
	if likeRes.Code != http.StatusOK {
		t.Fatalf("like status=%d body=%s", likeRes.Code, likeRes.Body.String())
	}
	payload := decodeJSON[map[string]any](t, likeRes)
	if payload["liked_user_id"] != targetID {
		t.Fatalf("liked user mismatch: got=%v want=%s", payload["liked_user_id"], targetID)
	}
	if payload["matched"].(bool) {
		t.Fatalf("expected first like not to create match")
	}
}

func TestMatchmakingMutualMatchCreation(t *testing.T) {
	env := setupIntegration(t)
	aID, aTokens := signupAndLogin(t, env.app, "mutual-a@example.com", "Password123")
	bID, bTokens := signupAndLogin(t, env.app, "mutual-b@example.com", "Password123")

	createProfile(t, env.app, aTokens.AccessToken, aID, 30, 90000, "Paris", map[string]bool{})
	createProfile(t, env.app, bTokens.AccessToken, bID, 30, 91000, "Paris", map[string]bool{})

	firstLike := requestJSON(t, env.app, http.MethodPost, "/matches/like", map[string]string{"liked_user_id": bID}, aTokens.AccessToken)
	if firstLike.Code != http.StatusOK {
		t.Fatalf("first like status=%d body=%s", firstLike.Code, firstLike.Body.String())
	}

	secondLike := requestJSON(t, env.app, http.MethodPost, "/matches/like", map[string]string{"liked_user_id": aID}, bTokens.AccessToken)
	if secondLike.Code != http.StatusOK {
		t.Fatalf("second like status=%d body=%s", secondLike.Code, secondLike.Body.String())
	}
	secondPayload := decodeJSON[map[string]any](t, secondLike)
	if !secondPayload["matched"].(bool) {
		t.Fatalf("expected reciprocal like to create mutual match")
	}

	relRes := requestJSON(t, env.app, http.MethodGet, "/matches/relationships", nil, aTokens.AccessToken)
	if relRes.Code != http.StatusOK {
		t.Fatalf("relationships status=%d body=%s", relRes.Code, relRes.Body.String())
	}
	var rel struct {
		MutualMatches []string `json:"mutual_matches"`
	}
	rel = decodeJSON[struct {
		MutualMatches []string `json:"mutual_matches"`
	}](t, relRes)
	if len(rel.MutualMatches) != 1 || rel.MutualMatches[0] != bID {
		t.Fatalf("unexpected mutual matches: %#v", rel.MutualMatches)
	}
}

func TestMatchmakingFilteredProfileSearch(t *testing.T) {
	env := setupIntegration(t)
	actorID, actorTokens := signupAndLogin(t, env.app, "filter-actor@example.com", "Password123")
	candidate1ID, candidate1Tokens := signupAndLogin(t, env.app, "filter-c1@example.com", "Password123")
	candidate2ID, candidate2Tokens := signupAndLogin(t, env.app, "filter-c2@example.com", "Password123")
	candidate3ID, candidate3Tokens := signupAndLogin(t, env.app, "filter-c3@example.com", "Password123")

	createProfile(t, env.app, actorTokens.AccessToken, actorID, 31, 110000, "Berlin", map[string]bool{})
	createProfile(t, env.app, candidate1Tokens.AccessToken, candidate1ID, 28, 100000, "Berlin", map[string]bool{})
	createProfile(t, env.app, candidate2Tokens.AccessToken, candidate2ID, 22, 100000, "Berlin", map[string]bool{})
	createProfile(t, env.app, candidate3Tokens.AccessToken, candidate3ID, 29, 100000, "Madrid", map[string]bool{})

	browseRes := requestJSON(t, env.app, http.MethodGet, "/matches/browse?min_age=25&max_age=30&location=Berlin&page=1&size=10", nil, actorTokens.AccessToken)
	if browseRes.Code != http.StatusOK {
		t.Fatalf("browse status=%d body=%s", browseRes.Code, browseRes.Body.String())
	}
	var payload struct {
		Items []struct {
			UserID   string `json:"user_id"`
			Age      int    `json:"age"`
			Location string `json:"location"`
		} `json:"items"`
	}
	payload = decodeJSON[struct {
		Items []struct {
			UserID   string `json:"user_id"`
			Age      int    `json:"age"`
			Location string `json:"location"`
		} `json:"items"`
	}](t, browseRes)
	if len(payload.Items) != 1 {
		t.Fatalf("expected exactly one filtered result, got=%d", len(payload.Items))
	}
	if payload.Items[0].UserID != candidate1ID {
		t.Fatalf("expected candidate1 only, got=%s", payload.Items[0].UserID)
	}
}
