package statements

import (
	"context"
	"errors"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements/explain"
)

type contestRegistry interface {
	Get(ctx context.Context, id domain.ContestID) (contests.ContestRecord, bool, error)
}

type Service struct {
	registry  contestRegistry
	providers map[string]Provider
	problems  *ProblemStore
	Explain   *explain.Service
}

func NewService(registry contestRegistry, providers map[string]Provider, problems *ProblemStore, labels explain.LabelResolver) *Service {
	cp := make(map[string]Provider, len(providers))
	for k, v := range providers {
		cp[k] = v
	}
	s := &Service{registry: registry, providers: cp, problems: problems}
	s.Explain = explain.New(s, labels)
	return s
}

func (s *Service) ExplainProblemStatement(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (domain.ProblemStatementExplain, error) {
	if s == nil || s.Explain == nil {
		return domain.ProblemStatementExplain{}, ErrProblemStatementNotFound
	}
	ex, err := s.Explain.Explain(ctx, contestID, problemID)
	if errors.Is(err, explain.ErrNotFound) {
		return domain.ProblemStatementExplain{}, ErrProblemStatementNotFound
	}
	return ex, err
}

func (s *Service) Fetch(ctx context.Context, contestID domain.ContestID) (Document, error) {
	if s == nil || s.registry == nil {
		return Document{}, fmt.Errorf("statements: service not configured")
	}
	record, ok, err := s.registry.Get(ctx, contestID)
	if err != nil {
		return Document{}, err
	}
	if !ok {
		return Document{}, contests.ErrContestNotFound
	}
	c := Contest{
		ID:         record.Contest.ID,
		ParallelID: record.Contest.ParallelID,
		Source:     record.Source,
	}
	key := resolveStatementSource(c)
	if key == "" {
		return Document{}, ErrNotAvailable
	}
	p, ok := s.providers[key]
	if !ok {
		return Document{}, ErrNotAvailable
	}
	return p.Fetch(ctx, c)
}
