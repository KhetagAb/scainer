package explain

import (
	"context"

	"scainer/internal/domain"
)

type Source interface {
	EnsureParsed(ctx context.Context, contestID domain.ContestID) error
	LoadStatement(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error)
}

type Service struct {
	src    Source
	labels LabelResolver
}

func New(src Source, labels LabelResolver) *Service {
	return &Service{src: src, labels: labels}
}

func (s *Service) Explain(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (domain.ProblemStatementExplain, error) {
	ps, err := s.resolveStatement(ctx, contestID, problemID)
	if err != nil {
		return domain.ProblemStatementExplain{}, err
	}
	// TODO: AI-explain — заменить parsed statement на LLM-объяснение.
	return domain.ProblemStatementExplain{
		Problem:   ps.Problem,
		Title:     ps.Title,
		Statement: ps.Statement,
	}, nil
}

func (s *Service) resolveStatement(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (domain.ProblemStatement, error) {
	if err := s.src.EnsureParsed(ctx, contestID); err != nil {
		return domain.ProblemStatement{}, err
	}

	letter, ok, err := s.labels.ProblemLabel(ctx, contestID, problemID)
	if err != nil {
		return domain.ProblemStatement{}, err
	}
	if !ok || letter == "" {
		return domain.ProblemStatement{}, ErrNotFound
	}
	return s.src.LoadStatement(contestID, domain.ProblemID(letter))
}
