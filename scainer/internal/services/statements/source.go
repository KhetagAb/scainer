package statements

import (
	"context"

	"scainer/internal/domain"
)

func (s *Service) EnsureParsed(ctx context.Context, contestID domain.ContestID) error {
	return s.EnsureContestStatements(ctx, contestID)
}

func (s *Service) LoadStatement(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error) {
	if s == nil || s.problems == nil {
		return domain.ProblemStatement{}, ErrProblemStatementNotFound
	}
	return s.problems.GetStatement(contestID, problemID)
}
