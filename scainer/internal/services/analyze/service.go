package analyze

import (
	"context"
	"sync"

	"scainer/internal/domain"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
)

type Service struct {
	pool    *jobs.Pool
	runner  *Runner
	mu      sync.Mutex
	running map[domain.ContestID]struct{}
}

func New(pool *jobs.Pool, runner *Runner) *Service {
	return &Service{
		pool:    pool,
		runner:  runner,
		running: make(map[domain.ContestID]struct{}),
	}
}

func (s *Service) Import(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		_, err := s.runner.Import(jobCtx, id)
		return err
	})
}

func (s *Service) ImportThenAnalyze(ctx context.Context, id domain.ContestID) (string, error) {
	return s.enqueue(ctx, id, func(jobCtx context.Context) error {
		if _, err := s.runner.Import(jobCtx, id); err != nil {
			return err
		}
		return s.runner.Analyze(jobCtx, id)
	})
}

func (s *Service) enqueue(
	ctx context.Context,
	id domain.ContestID,
	run func(context.Context) error,
) (string, error) {
	if _, err := lookup(ctx, s.runner.registry, id); err != nil {
		return "", err
	}
	if err := s.acquire(id); err != nil {
		return "", err
	}

	login, _ := auth.LoginFrom(ctx)
	jobID := s.pool.Submit(ctx, func(jobCtx context.Context) error {
		defer s.release(id)
		if login != "" {
			jobCtx = auth.WithLogin(jobCtx, login)
		}
		return run(jobCtx)
	})
	return jobID, nil
}

func (s *Service) acquire(id domain.ContestID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.running[id]; ok {
		return ErrJobRunning
	}
	s.running[id] = struct{}{}
	return nil
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
