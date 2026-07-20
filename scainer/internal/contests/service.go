package contests

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/importer"
	ejudgeapi "scainer/pkg/ejudge"
)

type Service struct {
	registry           ContestRegistry
	findingsStore      FindingsStore
	defaultJudgeSystem string
}

func NewService(registry ContestRegistry, findingsStore FindingsStore, defaultJudgeSystem string) *Service {
	return &Service{
		registry:           registry,
		findingsStore:      findingsStore,
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

	// ejudge: contest_id всегда = ID контеста в scainer; PostContests Source не заполняет.
	contestName := ""
	if source.Type == "ejudge" {
		cfgNode, err := ejudgeConfigNode(registration.ID)
		if err != nil {
			return Contest{}, err
		}
		source = &SourceSpec{Type: "ejudge", Config: *cfgNode}

		contestName, err = ejudgeContestName(ctx, registration.ID)
		if err != nil {
			return Contest{}, err
		}
	}

	if _, err := importer.Build(source.Type, &source.Config); err != nil {
		return Contest{}, err
	}

	contest := Contest{
		ID:               registration.ID,
		Name:             contestName,
		ParallelID:       registration.ParallelID,
		ParallelName:     registration.ParallelName,
		ExcludedProblems: registration.ExcludedProblems,
	}

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: *source}); err != nil {
		return Contest{}, fmt.Errorf("persist contest: %w", err)
	}

	return contest, nil
}

func (s *Service) SetParallel(ctx context.Context, id domain.ContestID, parallelID domain.ParallelID, parallelName string) error {
	record, err := get(ctx, s.registry, id)
	if err != nil {
		return err
	}
	contest := record.Contest
	contest.ParallelID = parallelID
	contest.ParallelName = parallelName

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
	if err := s.findingsStore.Delete(ctx, id); err != nil {
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

func ejudgeConfigNode(contestID domain.ContestID) (*yaml.Node, error) {
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

func ejudgeContestName(ctx context.Context, contestID domain.ContestID) (string, error) {
	cid, err := strconv.Atoi(string(contestID))
	if err != nil {
		return "", fmt.Errorf("invalid contest_id for ejudge: %s", contestID)
	}

	env, err := ejudgeapi.LoadEnv()
	if err != nil {
		return "", fmt.Errorf("load ejudge env: %w", err)
	}
	if env == nil || env.Client == nil {
		return "", errors.New("ejudge client not initialized")
	}

	info, err := env.Client.ContestStatus(ctx, cid)
	if err != nil {
		return "", fmt.Errorf("get contest status: %w", err)
	}
	return info.Name, nil
}
