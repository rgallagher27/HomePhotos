package scanner_test

import (
	"context"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/services/scanner"
)

func TestScheduler_OnStartup(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()
	createTestFile(t, dir, "startup.jpg")

	svc := scanner.New(repo, dir)
	sched := scanner.NewScheduler(svc, time.Hour, true)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.Start(ctx)
		close(done)
	}()

	// Wait a moment for the startup scan to complete
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	fps, err := repo.ListAllFingerprints(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(fps) != 1 {
		t.Errorf("photos = %d, want 1 (startup scan should have run)", len(fps))
	}
}

func TestScheduler_NoStartup(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	dir := t.TempDir()
	createTestFile(t, dir, "no-startup.jpg")

	svc := scanner.New(repo, dir)
	sched := scanner.NewScheduler(svc, time.Hour, false)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	fps, _ := repo.ListAllFingerprints(context.Background())
	if len(fps) != 0 {
		t.Errorf("photos = %d, want 0 (startup scan should not have run)", len(fps))
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)

	svc := scanner.New(repo, t.TempDir())
	sched := scanner.NewScheduler(svc, time.Hour, false)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.Start(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// OK — Start returned
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}
