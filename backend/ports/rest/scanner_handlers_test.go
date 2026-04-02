package rest

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPostScannerRun_Admin(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/scanner/run", "", token)
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status = %d, want 202", resp.StatusCode)
	}

	var body ScannerRunResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Status != "started" {
		t.Errorf("status = %q, want %q", body.Status, "started")
	}
}

func TestPostScannerRun_NonAdmin(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	// Register admin first, then viewer
	registerUser(t, handler, "admin", "password123")
	viewerToken := registerUser(t, handler, "viewer", "password123")

	resp := doRequest(t, handler, "POST", "/api/v1/scanner/run", "", viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestPostScannerRun_Unauthenticated(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	resp := doRequest(t, handler, "POST", "/api/v1/scanner/run", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestGetScannerStatus_Admin(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)
	token := registerUser(t, handler, "admin", "password123")

	resp := doRequest(t, handler, "GET", "/api/v1/scanner/status", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body ScannerStatusResponse
	json.NewDecoder(resp.Body).Decode(&body)
	if body.Status != "idle" {
		t.Errorf("status = %q, want %q", body.Status, "idle")
	}
}

func TestGetScannerStatus_NonAdmin(t *testing.T) {
	s := testServer(t, true)
	handler := buildHandler(t, s)

	registerUser(t, handler, "admin", "password123")
	viewerToken := registerUser(t, handler, "viewer", "password123")

	resp := doRequest(t, handler, "GET", "/api/v1/scanner/status", "", viewerToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}
