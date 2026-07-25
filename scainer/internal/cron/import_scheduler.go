package cron

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
)

type ImportScheduler struct {
	registry contests.ContestRegistry
	analyze  *analyze.Service
	interval time.Duration
}

func NewImportScheduler(registry contests.ContestRegistry, analyzeSvc *analyze.Service, interval time.Duration) *ImportScheduler {
	return &ImportScheduler{
		registry: registry,
		analyze:  analyzeSvc,
		interval: interval,
	}
}

func (s *ImportScheduler) Run(ctx context.Context) {
	if s.interval <= 0 {
		return
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *ImportScheduler) tick(ctx context.Context) {
	records, err := s.registry.List(ctx)
	if err != nil {
		log.Printf("import_cron: list contests: %v", err)
		return
	}

	var queued, skipped int
	for _, rec := range records {
		_, err := s.analyze.Import(ctx, rec.Contest.ID)
		if errors.Is(err, analyze.ErrJobRunning) {
			skipped++
			continue
		}
		if err != nil {
			log.Printf("import_cron: contest %s: %v", rec.Contest.ID, err)
			continue
		}
		queued++
	}
	if queued > 0 || skipped > 0 {
		log.Printf("import_cron: queued %d, skipped %d (running)", queued, skipped)
	}
}

func StartImportCron(ctx context.Context, registry contests.ContestRegistry, analyzeSvc *analyze.Service, interval time.Duration) (func(), error) {
	if interval <= 0 {
		return nil, fmt.Errorf("import_cron.interval must be > 0")
	}
	sched := NewImportScheduler(registry, analyzeSvc, interval)
	runCtx, cancel := context.WithCancel(ctx)
	go sched.Run(runCtx)
	return cancel, nil
}
