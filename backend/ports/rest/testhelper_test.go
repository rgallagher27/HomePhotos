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
	"github.com/rgallagher/homephotos/services/cache"
	"github.com/rgallagher/homephotos/services/scanner"
)

type testEnv struct {
	server    *Server
	photos    *sqlite.PhotoRepository
	sourceDir string
	cacheDir  string
}

func newTestEnv(t *testing.T, registrationOpen bool) *testEnv {
	t.Helper()

	db := setupTestDB(t)
	tokens := auth.NewTokenService("test-secret", time.Hour)
	userRepo := sqlite.NewUserRepository(db)
	authSvc := auth.New(userRepo, tokens, 4, registrationOpen)
	photoRepo := sqlite.NewPhotoRepository(db)
	sourceDir := t.TempDir()
	cacheDir := t.TempDir()
	scannerSvc := scanner.New(photoRepo, sourceDir)
	cacheSvc := cache.New(photoRepo, sourceDir, cacheDir)

	return &testEnv{
		server:    NewServer(db, authSvc, tokens, userRepo, photoRepo, scannerSvc, cacheSvc),
		photos:    photoRepo,
		sourceDir: sourceDir,
		cacheDir:  cacheDir,
	}
}

// testServer creates a Server backed by an in-memory SQLite database with migrations applied.
func testServer(t *testing.T, registrationOpen bool) *Server {
	t.Helper()
	return newTestEnv(t, registrationOpen).server
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

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
