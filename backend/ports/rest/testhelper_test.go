package rest

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/services/auth"
	"github.com/rgallagher/homephotos/services/scanner"
)

// testServer creates a Server backed by an in-memory SQLite database with migrations applied.
func testServer(t *testing.T, registrationOpen bool) *Server {
	t.Helper()

	db := setupTestDB(t)
	tokens := auth.NewTokenService("test-secret", time.Hour)
	userRepo := sqlite.NewUserRepository(db)
	authSvc := auth.New(userRepo, tokens, 4, registrationOpen) // cost 4 for fast tests
	photoRepo := sqlite.NewPhotoRepository(db)
	scannerSvc := scanner.New(photoRepo, t.TempDir())

	return NewServer(db, authSvc, tokens, userRepo, photoRepo, scannerSvc)
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)

	migrationsDir := filepath.Join("..", "..", "database", "sqlite", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		data, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			t.Fatalf("exec %s: %v", f, err)
		}
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// registerUser is a test helper that registers a user and returns the auth token.
func registerUser(t *testing.T, handler http.Handler, username, password string) string {
	t.Helper()
	body := `{"username":"` + username + `","password":"` + password + `"}`
	resp := doRequest(t, handler, "POST", "/api/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status = %d, want 201", username, resp.StatusCode)
	}
	return extractToken(t, resp)
}

// loginUser logs in and returns the auth token.
func loginUser(t *testing.T, handler http.Handler, username, password string) string {
	t.Helper()
	body := `{"username":"` + username + `","password":"` + password + `"}`
	resp := doRequest(t, handler, "POST", "/api/v1/auth/login", body, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s: status = %d, want 200", username, resp.StatusCode)
	}
	return extractToken(t, resp)
}
