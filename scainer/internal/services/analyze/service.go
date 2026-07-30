package analyze

import (
	"context"
	"errors"
	"sync"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/refresh"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
)

const pendingJobID = "__pending__"

type Service struct {
	pool     *jobs.Pool
	refresh  *refresh.Orchestrator
	registry contests.ContestRegistry
	mu       sync.Mutex
	running  map[domain.ContestID]string
}

func New(pool *jobs.Pool, refreshOrch *refresh.Orchestrator, registry contests.ContestRegistry) *Service {
	return &Service{
		pool:     pool,
		refresh:  refreshOrch,
		registry: registry,
		running:  make(map[domain.ContestID]string),
	}
}

func (s *Service) Import(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		_, err := s.refresh.Import(jobCtx, id)
		return err
	})
}

func (s *Service) Analyze(ctx context.Context, id domain.ContestID) (string, error) {
	if err := RequireImported(ctx, s.registry, id); err != nil {
		return "", err
	}
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		return s.refresh.Analyze(WithManual(jobCtx), id)
	})
}

func (s *Service) Sync(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		return s.refresh.Sync(jobCtx, id)
	})
}

func (s *Service) SyncManual(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		return s.refresh.Sync(WithManual(jobCtx), id)
	})
}

func (s *Service) ResyncManual(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		return s.refresh.Resync(WithManual(jobCtx), id)
	})
}

func (s *Service) enqueue(
	ctx context.Context,
	id domain.ContestID,
	run func(context.Context) error,
) (string, error) {
	if _, err := lookup(ctx, s.registry, id); err != nil {
		return "", err
	}

	if existing, ok, err := s.existingJobID(id); err != nil {
		return "", err
	} else if ok {
		return existing, nil
	}

	login, _ := auth.LoginFrom(ctx)
	jobID := s.pool.Submit(ctx, func(jobCtx context.Context) error {
		defer s.release(id)
		if login != "" {
			jobCtx = auth.WithLogin(jobCtx, login)
		}
		return run(jobCtx)
	})
	s.setJobID(id, jobID)
	return jobID, nil
}

// existingJobID возвращает jobId уже идущего import/analyze для контеста.
// Если другой goroutine только резервирует слот — ждём до 2 с.
func (s *Service) existingJobID(id domain.ContestID) (string, bool, error) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		jobID, ok := s.running[id]
		s.mu.Unlock()
		if !ok {
			if reserved := s.reserve(id); !reserved {
				continue
			}
			return "", false, nil
		}
		if jobID != pendingJobID {
			return jobID, true, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", false, errors.New("timeout waiting for contest job reservation")
}

func (s *Service) reserve(id domain.ContestID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.running[id]; ok {
		return false
	}
	s.running[id] = pendingJobID
	return true
}

func (s *Service) setJobID(id domain.ContestID, jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running[id] = jobID
}

func (s *Service) release(id domain.ContestID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, id)
}

func (s *Service) JobStatus(id string) (jobs.State, bool) {
	return s.pool.Get(id)
}

func (s *Service) SubscribeJob(id string) (<-chan jobs.State, func(), bool) {
	return s.pool.Subscribe(id)
}
