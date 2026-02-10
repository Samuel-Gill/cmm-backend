package integration_test

import (
	"net/http"
	"testing"
)

func TestProfileCreateAndUpdate(t *testing.T) {
	env := setupIntegration(t)
	userID, tokens := signupAndLogin(t, env.app, "profile-owner@example.com", "Password123")

	createPayload := map[string]any{
		"user_id":           userID,
		"age":               29,
		"profession":        "Backend Engineer",
		"education":         "Bachelor",
		"income":            98000,
		"residency_status":  "permanent_resident",
		"location":          "Berlin",
		"marital_status":    "single",
		"description":       "initial",
		"hide_contact_info": true,
		"hide_address":      false,
		"hide_income":       true,
		"hide_visa_status":  true,
	}
	createRes := requestJSON(t, env.app, http.MethodPost, "/profiles/", createPayload, tokens.AccessToken)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createRes.Code, createRes.Body.String())
	}
	created := decodeJSON[map[string]any](t, createRes)
	if int(created["age"].(float64)) != 29 {
		t.Fatalf("unexpected age: %v", created["age"])
	}

	updatePayload := map[string]any{
		"age":               31,
		"profession":        "Staff Engineer",
		"education":         "Master",
		"income":            130000,
		"residency_status":  "citizen",
		"location":          "Munich",
		"marital_status":    "single",
		"description":       "updated",
		"hide_contact_info": true,
		"hide_address":      true,
		"hide_income":       true,
		"hide_visa_status":  false,
	}
	updateRes := requestJSON(t, env.app, http.MethodPut, "/profiles/"+userID, updatePayload, tokens.AccessToken)
	if updateRes.Code != http.StatusOK {
		t.Fatalf("update profile status=%d body=%s", updateRes.Code, updateRes.Body.String())
	}
	updated := decodeJSON[map[string]any](t, updateRes)
	if updated["location"] != "Munich" {
		t.Fatalf("expected location Munich got=%v", updated["location"])
	}
	if !updated["hide_address"].(bool) {
		t.Fatalf("expected hide_address true")
	}
}

func TestProfilePrivacyFlagsInGetAndListResponses(t *testing.T) {
	env := setupIntegration(t)
	userID, tokens := signupAndLogin(t, env.app, "profile-privacy@example.com", "Password123")

	createProfile(t, env.app, tokens.AccessToken, userID, 27, 85000, "Tokyo", map[string]bool{
		"hide_contact_info": true,
		"hide_address":      true,
		"hide_income":       true,
		"hide_visa_status":  false,
	})

	getRes := requestJSON(t, env.app, http.MethodGet, "/profiles/"+userID, nil, "")
	if getRes.Code != http.StatusOK {
		t.Fatalf("get profile status=%d body=%s", getRes.Code, getRes.Body.String())
	}
	getPayload := decodeJSON[map[string]any](t, getRes)
	if !getPayload["hide_contact_info"].(bool) || !getPayload["hide_address"].(bool) || !getPayload["hide_income"].(bool) {
		t.Fatalf("expected privacy flags to be returned as configured")
	}

	listRes := requestJSON(t, env.app, http.MethodGet, "/profiles/?page=1&size=10", nil, "")
	if listRes.Code != http.StatusOK {
		t.Fatalf("list profile status=%d body=%s", listRes.Code, listRes.Body.String())
	}
	var listPayload struct {
		Items []map[string]any `json:"items"`
	}
	listPayload = decodeJSON[struct {
		Items []map[string]any `json:"items"`
	}](t, listRes)
	if len(listPayload.Items) == 0 {
		t.Fatalf("expected at least one profile in list")
	}
	if !listPayload.Items[0]["hide_income"].(bool) {
		t.Fatalf("expected hide_income flag in list response")
	}
}

func TestProfileOwnerOnlyCanUpdate(t *testing.T) {
	env := setupIntegration(t)
	ownerID, ownerTokens := signupAndLogin(t, env.app, "profile-owner-update@example.com", "Password123")
	_, attackerTokens := signupAndLogin(t, env.app, "profile-attacker@example.com", "Password123")

	createProfile(t, env.app, ownerTokens.AccessToken, ownerID, 33, 92000, "London", map[string]bool{})

	res := requestJSON(t, env.app, http.MethodPut, "/profiles/"+ownerID, map[string]any{
		"age":               34,
		"profession":        "Architect",
		"education":         "Master",
		"income":            120000,
		"residency_status":  "citizen",
		"location":          "London",
		"marital_status":    "single",
		"description":       "attempted update",
		"hide_contact_info": false,
		"hide_address":      false,
		"hide_income":       false,
		"hide_visa_status":  false,
	}, attackerTokens.AccessToken)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status=%d body=%s", res.Code, res.Body.String())
	}
}
