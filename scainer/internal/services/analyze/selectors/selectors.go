package selectors

import (
	"context"

	"scainer/internal/domain"
)

type Store interface {
	ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error)
}

type Selector[U domain.Unit] interface {
	Select(ctx context.Context, s Store) ([]U, error)
}
