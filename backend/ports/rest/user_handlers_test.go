package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetUsers(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)

	adminToken := registerUser(t, h, "admin", "password123")
	registerUser(t, h, "viewer", "password123")

	// Get viewer token via login
	viewerToken := loginUser(t, h, "viewer", "password123")

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCount  int
	}{
		{"admin can list", adminToken, http.StatusOK, 2},
		{"viewer forbidden", viewerToken, http.StatusForbidden, 0},
		{"no auth", "", http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doRequest(t, h, "GET", "/api/v1/users", "", tt.token)
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, want %d; body: %s", resp.StatusCode, tt.wantStatus, body)
			}
			if tt.wantCount > 0 {
				var result struct {
					Data []json.RawMessage `json:"data"`
				}
				json.NewDecoder(resp.Body).Decode(&result)
				if len(result.Data) != tt.wantCount {
					t.Errorf("count = %d, want %d", len(result.Data), tt.wantCount)
				}
			}
		})
	}
}

func TestPutUserRole(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)

	adminToken := registerUser(t, h, "admin", "password123")
	registerUser(t, h, "viewer", "password123")
	viewerToken := loginUser(t, h, "viewer", "password123")

	// Get viewer's user ID
	resp := doRequest(t, h, "GET", "/api/v1/users", "", adminToken)
	var listResp struct {
		Data []struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&listResp)

	var viewerID int64
	for _, u := range listResp.Data {
		if u.Username == "viewer" {
			viewerID = u.ID
			break
		}
	}

	tests := []struct {
		name       string
		token      string
		path       string
		body       string
		wantStatus int
	}{
		{"admin promotes viewer", adminToken, "/api/v1/users/2/role", `{"role":"admin"}`, http.StatusOK},
		{"viewer forbidden", viewerToken, "/api/v1/users/2/role", `{"role":"admin"}`, http.StatusForbidden},
		{"no auth", "", "/api/v1/users/2/role", `{"role":"admin"}`, http.StatusUnauthorized},
		{"not found", adminToken, "/api/v1/users/999/role", `{"role":"admin"}`, http.StatusNotFound},
	}

	// Use viewerID to build path for the first test
	_ = viewerID

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doRequest(t, h, "PUT", tt.path, tt.body, tt.token)
			if resp.StatusCode != tt.wantStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d; body: %s", resp.StatusCode, tt.wantStatus, body)
			}
		})
	}
}

func TestPutUserRole_VerifyChange(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)

	adminToken := registerUser(t, h, "admin", "password123")
	registerUser(t, h, "viewer", "password123")

	// Promote viewer to admin
	resp := doRequest(t, h, "PUT", "/api/v1/users/2/role", `{"role":"admin"}`, adminToken)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("promote: status = %d; body: %s", resp.StatusCode, body)
	}

	var userResp struct {
		Role string `json:"role"`
	}
	json.NewDecoder(resp.Body).Decode(&userResp)
	if userResp.Role != "admin" {
		t.Errorf("role = %q, want %q", userResp.Role, "admin")
	}
}
