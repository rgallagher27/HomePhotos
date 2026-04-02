package scanner_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/services/scanner"
)

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

func createTestFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test content"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestService_Run_NewFiles(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "photo1.jpg")
	createTestFile(t, dir, "subdir/photo2.arw")
	createTestFile(t, dir, "photo3.png")

	svc := scanner.New(repo, dir)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}

	fps, err := repo.ListAllFingerprints(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(fps) != 3 {
		t.Errorf("photos = %d, want 3", len(fps))
	}

	// Verify a photo was created correctly
	p, err := repo.GetByFilePath(context.Background(), "photo1.jpg")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if p.Format != "jpg" {
		t.Errorf("format = %q, want %q", p.Format, "jpg")
	}
	if p.CacheStatus != photo.CacheStatusPending {
		t.Errorf("cache_status = %q, want %q", p.CacheStatus, photo.CacheStatusPending)
	}
}

func TestService_Run_Incremental(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "photo.jpg")

	svc := scanner.New(repo, dir)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// Get the fingerprint after first scan
	p1, _ := repo.GetByFilePath(context.Background(), "photo.jpg")
	fp1 := p1.Fingerprint

	// Run again — should be a no-op
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("second run: %v", err)
	}

	p2, _ := repo.GetByFilePath(context.Background(), "photo.jpg")
	if p2.Fingerprint != fp1 {
		t.Errorf("fingerprint changed on no-op scan")
	}
}

func TestService_Run_ChangedFile(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "photo.jpg")

	svc := scanner.New(repo, dir)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("first run: %v", err)
	}

	p1, _ := repo.GetByFilePath(context.Background(), "photo.jpg")
	fp1 := p1.Fingerprint

	// Modify the file (change size and mtime)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("modified content here"), 0o644)

	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("second run: %v", err)
	}

	p2, _ := repo.GetByFilePath(context.Background(), "photo.jpg")
	if p2.Fingerprint == fp1 {
		t.Error("fingerprint should have changed after file modification")
	}
}

func TestService_Run_OrphanedFile(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "keep.jpg")
	createTestFile(t, dir, "remove.jpg")

	svc := scanner.New(repo, dir)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// Remove one file
	os.Remove(filepath.Join(dir, "remove.jpg"))

	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("second run: %v", err)
	}

	fps, _ := repo.ListAllFingerprints(context.Background())
	if len(fps) != 1 {
		t.Errorf("photos = %d, want 1", len(fps))
	}
	if _, exists := fps["keep.jpg"]; !exists {
		t.Error("expected keep.jpg to remain")
	}
}

func TestService_Run_UnsupportedExtension(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "readme.txt")
	createTestFile(t, dir, "photo.jpg")

	svc := scanner.New(repo, dir)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}

	fps, _ := repo.ListAllFingerprints(context.Background())
	if len(fps) != 1 {
		t.Errorf("photos = %d, want 1 (txt should be ignored)", len(fps))
	}
}

func TestService_Run_ConcurrencyGuard(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	createTestFile(t, dir, "test.jpg")

	svc := scanner.New(repo, dir)

	// Start a scan in a goroutine with a context we control
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run one scan to completion first, then start a blocking scan
	// by running two concurrent scans with enough work
	// Simpler: run scan, while it's running try another
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		// Create many files to slow the scan
		for i := 0; i < 100; i++ {
			createTestFile(t, dir, filepath.Join("batch", fmt.Sprintf("file%d.jpg", i)))
		}
		close(started)
		done <- svc.Run(ctx)
	}()

	<-started
	// Try to run concurrently — the first scan may or may not have started yet
	// so we retry a few times
	var gotInProgress bool
	for i := 0; i < 20; i++ {
		err := svc.Run(context.Background())
		if err == scanner.ErrScanInProgress {
			gotInProgress = true
			break
		}
		time.Sleep(time.Millisecond)
	}

	cancel()
	<-done

	if !gotInProgress {
		t.Error("expected ErrScanInProgress error")
	}
}

func TestService_Status(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()

	svc := scanner.New(repo, dir)

	// Initially idle
	status := svc.Status()
	if status.State != "idle" {
		t.Errorf("state = %q, want idle", status.State)
	}
	if status.StartedAt != nil {
		t.Error("expected nil started_at when idle")
	}

	// After scan completes, should be idle again
	createTestFile(t, dir, "test.jpg")
	svc.Run(context.Background())

	status = svc.Status()
	if status.State != "idle" {
		t.Errorf("state = %q, want idle after scan", status.State)
	}
}
