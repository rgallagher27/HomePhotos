package scanner

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	scanner  *Service
	interval time.Duration
	onStart  bool
}

func NewScheduler(scanner *Service, interval time.Duration, onStart bool) *Scheduler {
	return &Scheduler{
		scanner:  scanner,
		interval: interval,
		onStart:  onStart,
	}
}

// Start blocks until ctx is cancelled. It runs an initial scan if onStart is true,
// then runs periodic scans at the configured interval.
func (s *Scheduler) Start(ctx context.Context) {
	if s.onStart {
		slog.InfoContext(ctx, "running startup scan")
		if err := s.scanner.Run(ctx); err != nil {
			slog.WarnContext(ctx, "startup scan failed", "error", err)
		}
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			slog.InfoContext(ctx, "running scheduled scan")
			if err := s.scanner.Run(ctx); err != nil {
				slog.WarnContext(ctx, "scheduled scan failed", "error", err)
			}
		}
	}
}
