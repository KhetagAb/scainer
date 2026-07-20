package analyze

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/jobs"
)

type Service struct {
	pool   *jobs.Pool
	runner *Runner
}

func New(pool *jobs.Pool, runner *Runner) *Service {
	return &Service{pool: pool, runner: runner}
}

func (s *Service) Submit(ctx context.Context, id domain.ContestID) (string, error) {
	if _, err := lookup(ctx, s.runner.registry, id); err != nil {
		return "", err
	}
	jobID := s.pool.Submit(ctx, func(jobCtx context.Context) error {
		_, err := s.runner.Run(jobCtx, id)
		return err
	})
	return jobID, nil
}

func (s *Service) JobStatus(id string) (jobs.State, bool) {
	return s.pool.Get(id)
}

func (s *Service) SubscribeJob(id string) (<-chan jobs.State, func(), bool) {
	return s.pool.Subscribe(id)
}
