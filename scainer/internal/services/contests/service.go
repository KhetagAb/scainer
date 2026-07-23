package contests

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/importer"
)

type Service struct {
	registry           ContestRegistry
	findingsRepo       FindingsRepository
	defaultJudgeSystem string
}

func NewService(registry ContestRegistry, findingsRepo FindingsRepository, defaultJudgeSystem string) *Service {
	return &Service{
		registry:           registry,
		findingsRepo:       findingsRepo,
		defaultJudgeSystem: defaultJudgeSystem,
	}
}

func (s *Service) Register(ctx context.Context, registration Registration) (Contest, error) {
	if _, err := get(ctx, s.registry, registration.ID); err == nil {
		return Contest{}, ErrDuplicateContest
	} else if !errors.Is(err, ErrContestNotFound) {
		return Contest{}, err
	}

	source := registration.Source
	if source == nil {
		source = &SourceSpec{
			Type:   s.defaultJudgeSystem,
			Config: yaml.Node{},
		}
	}

	// ejudge: contest_id = ID в scainer; имя подтянется при import (ContestStatus).
	if source.Type == "ejudge" {
		cfgNode, err := ejudgeSourceConfig(registration.ID)
		if err != nil {
			return Contest{}, err
		}
		source = &SourceSpec{Type: "ejudge", Config: *cfgNode}
	}

	if _, err := importer.Build(source.Type, &source.Config); err != nil {
		return Contest{}, err
	}

	contest := Contest{
		ID:               registration.ID,
		ParallelID:       registration.ParallelID,
		ExcludedProblems: registration.ExcludedProblems,
	}

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: *source}); err != nil {
		return Contest{}, fmt.Errorf("persist contest: %w", err)
	}

	return contest, nil
}

func (s *Service) SetParallel(ctx context.Context, id domain.ContestID, parallelID string) error {
	record, err := get(ctx, s.registry, id)
	if err != nil {
		return err
	}
	contest := record.Contest
	contest.ParallelID = parallelID

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: record.Source}); err != nil {
		return fmt.Errorf("persist contest: %w", err)
	}
	return nil
}

func (s *Service) RemoveContest(ctx context.Context, id domain.ContestID) error {
	if _, err := get(ctx, s.registry, id); err != nil {
		return err
	}
	if err := s.registry.Delete(ctx, id); err != nil {
		return fmt.Errorf("persist contest removal: %w", err)
	}
	if err := s.findingsRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("persist findings removal: %w", err)
	}
	return nil
}

func (s *Service) SetExcludedProblems(ctx context.Context, id domain.ContestID, problems []domain.ProblemID) error {
	record, err := get(ctx, s.registry, id)
	if err != nil {
		return err
	}
	contest := record.Contest
	contest.ExcludedProblems = problems

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: record.Source}); err != nil {
		return fmt.Errorf("persist contest: %w", err)
	}
	return nil
}

func ejudgeSourceConfig(contestID domain.ContestID) (*yaml.Node, error) {
	cid, err := strconv.Atoi(string(contestID))
	if err != nil {
		return nil, fmt.Errorf("invalid contest_id for ejudge: %s", contestID)
	}
	raw, err := yaml.Marshal(map[string]int{"contest_id": cid})
	if err != nil {
		return nil, fmt.Errorf("marshal ejudge config: %w", err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal ejudge config: %w", err)
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &doc, nil
}
