package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/pkg/auth"
)

type ImportScheduler struct {
	registry contests.ContestRegistry
	analyze  *analyze.Service
	interval time.Duration
	login    string
}

func NewImportScheduler(registry contests.ContestRegistry, analyzeSvc *analyze.Service, interval time.Duration, login string) *ImportScheduler {
	return &ImportScheduler{
		registry: registry,
		analyze:  analyzeSvc,
		interval: interval,
		login:    login,
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
	ctx = auth.WithLogin(ctx, s.login)

	records, err := s.registry.List(ctx)
	if err != nil {
		log.Printf("import_cron: list contests: %v", err)
		return
	}

	log.Printf("import_cron: tick contests=%d login=%s", len(records), s.login)

	var queued, skipped int
	for _, rec := range records {
		_, err := s.analyze.Sync(ctx, rec.Contest.ID)
		if err != nil {
			skipped++
			log.Printf("import_cron: contest=%s err=%v", rec.Contest.ID, err)
			continue
		}
		queued++
	}
	log.Printf("import_cron: queued=%d skipped=%d", queued, skipped)
}

func StartImportCron(ctx context.Context, registry contests.ContestRegistry, analyzeSvc *analyze.Service, interval time.Duration, login string) (func(), error) {
	if interval <= 0 {
		return nil, fmt.Errorf("import_cron.interval must be > 0")
	}
	if login == "" {
		return nil, fmt.Errorf("import_cron: masterlogin обязателен")
	}
	sched := NewImportScheduler(registry, analyzeSvc, interval, login)
	runCtx, cancel := context.WithCancel(ctx)
	go sched.Run(runCtx)
	return cancel, nil
}
