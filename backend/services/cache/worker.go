package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/rgallagher/homephotos/domain/photo"
)

// WorkerPool processes pending photos in the background.
type WorkerPool struct {
	cache      *Service
	photos     photo.Repository
	numWorkers int
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(svc *Service, numWorkers int) *WorkerPool {
	return &WorkerPool{
		cache:      svc,
		photos:     svc.Photos(),
		numWorkers: numWorkers,
	}
}

// Start launches the worker pool. Blocks until ctx is cancelled.
func (wp *WorkerPool) Start(ctx context.Context) {
	slog.Info("starting cache worker pool", "workers", wp.numWorkers)

	for {
		select {
		case <-ctx.Done():
			slog.Info("cache worker pool stopped")
			return
		default:
		}

		pending, err := wp.photos.ListPending(ctx, wp.numWorkers)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("cache worker pool stopped")
				return
			}
			slog.Warn("list pending for cache", "error", err)
			sleepOrCancel(ctx, 5*time.Second)
			continue
		}

		if len(pending) == 0 {
			sleepOrCancel(ctx, 5*time.Second)
			continue
		}

		slog.Info("processing pending photos", "count", len(pending))
		wp.processBatch(ctx, pending)
	}
}

func (wp *WorkerPool) processBatch(ctx context.Context, batch []photo.Photo) {
	var wg sync.WaitGroup
	for _, p := range batch {
		wg.Add(1)
		go func(p photo.Photo) {
			defer wg.Done()
			if err := wp.cache.Generate(ctx, &p); err != nil {
				slog.Warn("worker generate failed", "photo_id", p.ID, "error", err)
			}
		}(p)
	}
	wg.Wait()
}

func sleepOrCancel(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}
