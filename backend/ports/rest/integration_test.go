package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestIntegration_AuthFlow(t *testing.T) {
	s := testServer(t, true)
	h := buildHandler(t, s)

	// 1. Health check works without auth
	resp := doRequest(t, h, "GET", "/health", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health: status = %d, want 200", resp.StatusCode)
	}

	// 2. Register first user → admin
	resp = doRequest(t, h, "POST", "/api/v1/auth/register",
		`{"username":"admin","password":"password123","email":"admin@example.com"}`, "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("register admin: status = %d; body: %s", resp.StatusCode, body)
	}
	var adminAuth struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
		Token    string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&adminAuth)

	if adminAuth.Role != "admin" {
		t.Errorf("first user role = %q, want admin", adminAuth.Role)
	}
	if adminAuth.Token == "" {
		t.Fatal("expected token for admin")
	}

	// 3. GET /auth/me with admin token
	resp = doRequest(t, h, "GET", "/api/v1/auth/me", "", adminAuth.Token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("me: status = %d; body: %s", resp.StatusCode, body)
	}
	var me struct {
		Username string `json:"username"`
		Role     string `json:"role"`
		Email    string `json:"email"`
	}
	json.NewDecoder(resp.Body).Decode(&me)
	if me.Username != "admin" {
		t.Errorf("me username = %q, want admin", me.Username)
	}
	if me.Email != "admin@example.com" {
		t.Errorf("me email = %q, want admin@example.com", me.Email)
	}

	// 4. Register second user → viewer
	resp = doRequest(t, h, "POST", "/api/v1/auth/register",
		`{"username":"viewer","password":"password123"}`, "")
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("register viewer: status = %d; body: %s", resp.StatusCode, body)
	}
	var viewerAuth struct {
		ID    int64  `json:"id"`
		Role  string `json:"role"`
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&viewerAuth)
	if viewerAuth.Role != "viewer" {
		t.Errorf("second user role = %q, want viewer", viewerAuth.Role)
	}

	// 5. Login as viewer
	viewerToken := loginUser(t, h, "viewer", "password123")
	if viewerToken == "" {
		t.Fatal("expected token for viewer")
	}

	// 6. Admin lists users → sees both
	resp = doRequest(t, h, "GET", "/api/v1/users", "", adminAuth.Token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("list users: status = %d; body: %s", resp.StatusCode, body)
	}
	var listResp struct {
		Data []struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&listResp)
	if len(listResp.Data) != 2 {
		t.Fatalf("user count = %d, want 2", len(listResp.Data))
	}

	// 7. Viewer can't list users
	resp = doRequest(t, h, "GET", "/api/v1/users", "", viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("viewer list: status = %d, want 403", resp.StatusCode)
	}

	// 8. Admin promotes viewer to admin
	resp = doRequest(t, h, "PUT", "/api/v1/users/2/role",
		`{"role":"admin"}`, adminAuth.Token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("promote: status = %d; body: %s", resp.StatusCode, body)
	}
	var promoted struct {
		Role string `json:"role"`
	}
	json.NewDecoder(resp.Body).Decode(&promoted)
	if promoted.Role != "admin" {
		t.Errorf("promoted role = %q, want admin", promoted.Role)
	}

	// 9. Viewer's old token still has viewer role — needs to re-login
	resp = doRequest(t, h, "GET", "/api/v1/users", "", viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("old token should still be viewer: status = %d, want 403", resp.StatusCode)
	}

	// 10. Re-login as promoted user → gets admin token
	newToken := loginUser(t, h, "viewer", "password123")
	resp = doRequest(t, h, "GET", "/api/v1/users", "", newToken)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("promoted viewer with new token: status = %d; body: %s", resp.StatusCode, body)
	}
}
