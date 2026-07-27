package statements

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

type contestRegistry interface {
	Get(ctx context.Context, id domain.ContestID) (contests.ContestRecord, bool, error)
}

type Service struct {
	registry  contestRegistry
	providers map[string]Provider
}

func NewService(registry contestRegistry, providers map[string]Provider) *Service {
	cp := make(map[string]Provider, len(providers))
	for k, v := range providers {
		cp[k] = v
	}
	return &Service{registry: registry, providers: cp}
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
