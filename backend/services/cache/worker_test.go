package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/rgallagher/homephotos/database/sqlite"
	"github.com/rgallagher/homephotos/domain/photo"
	"github.com/rgallagher/homephotos/services/cache"
	"github.com/rgallagher/homephotos/services/imaging"
)

func TestWorkerPool_ProcessesPending(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)
	ctx := context.Background()

	sourceDir := t.TempDir()
	cacheDir := t.TempDir()

	// Create source files and photo records
	names := []string{"a", "b", "c"}
	var ids []int64
	for _, name := range names {
		createSourceJPEG(t, sourceDir, name+".jpg", 400, 300)
		p := newTestPhoto(name+".jpg", "fp-"+name)
		created, err := repo.Create(ctx, p)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		ids = append(ids, created.ID)
	}

	svc := cache.New(repo, sourceDir, cacheDir)
	wp := cache.NewWorkerPool(svc, 2)

	// Run with a timeout context
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		wp.Start(runCtx)
		close(done)
	}()

	// Poll until all photos are cached
	deadline := time.After(10 * time.Second)
	for {
		select {
		case <-deadline:
			cancel()
			t.Fatal("timed out waiting for worker pool to process photos")
		default:
		}

		pending, _ := repo.ListPending(ctx, 10)
		if len(pending) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	cancel()
	<-done

	// Verify all photos are cached
	for _, id := range ids {
		got, _ := repo.GetByID(ctx, id)
		if got.CacheStatus != photo.CacheStatusCached {
			t.Errorf("photo %d: status = %q, want cached", id, got.CacheStatus)
		}
		if !svc.Has(got.Fingerprint, imaging.SizeThumb) {
			t.Errorf("photo %d: thumb not found", id)
		}
		if !svc.Has(got.Fingerprint, imaging.SizePreview) {
			t.Errorf("photo %d: preview not found", id)
		}
	}
}

func TestWorkerPool_StopsOnContextCancel(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewPhotoRepository(db)

	svc := cache.New(repo, t.TempDir(), t.TempDir())
	wp := cache.NewWorkerPool(svc, 2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	done := make(chan struct{})
	go func() {
		wp.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Good — pool stopped
	case <-time.After(5 * time.Second):
		t.Fatal("worker pool did not stop after context cancellation")
	}
}
