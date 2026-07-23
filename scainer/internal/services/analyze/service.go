package analyze

import (
	"context"

	"scainer/internal/domain"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
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
	login, _ := auth.LoginFrom(ctx)
	jobID := s.pool.Submit(ctx, func(jobCtx context.Context) error {
		if login != "" {
			jobCtx = auth.WithLogin(jobCtx, login)
		}
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
