package analyze

import (
	"context"
	"fmt"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/detect"
	"scainer/internal/services/importer"
	"scainer/internal/services/scoring"
	"scainer/pkg/progress"
)

type Result struct {
	ImportedCount  int
	LastImportedAt time.Time
}

type Runner struct {
	registry        contests.ContestRegistry
	submissionStore contests.SubmissionStore
	findingsRepo    contests.FindingsRepository
	scorer          scoring.Scorer
	pipeline        Pipeline
}

func NewRunner(
	registry contests.ContestRegistry,
	submissionStore contests.SubmissionStore,
	findingsRepo contests.FindingsRepository,
	scorer scoring.Scorer,
	pipeline Pipeline,
) *Runner {
	return &Runner{
		registry:        registry,
		submissionStore: submissionStore,
		findingsRepo:    findingsRepo,
		scorer:          scorer,
		pipeline:        pipeline,
	}
}

func (r *Runner) Import(ctx context.Context, id domain.ContestID) (Result, error) {
	progress.Report(ctx, progress.Event{Phase: "importing"})

	record, err := lookup(ctx, r.registry, id)
	if err != nil {
		return Result{}, err
	}

	contestImporter, err := importer.Build(record.Source.Type, &record.Source.Config)
	if err != nil {
		return Result{}, fmt.Errorf("build importer: %w", err)
	}

	imported, err := contestImporter.Import(ctx, r.submissionStore)
	if err != nil {
		return Result{}, err
	}
	if err := r.submissionStore.Put(ctx, imported.Submissions); err != nil {
		return Result{}, err
	}

	now := time.Now().UTC().Truncate(time.Second)
	contest := record.Contest
	contest.LastImportedAt = &now
	contest.Name = imported.ContestName
	if err := r.registry.Put(ctx, contests.ContestRecord{Contest: contest, Source: record.Source}); err != nil {
		return Result{}, fmt.Errorf("persist lastImportedAt: %w", err)
	}

	return Result{ImportedCount: len(imported.Submissions), LastImportedAt: now}, nil
}

func (r *Runner) Analyze(ctx context.Context, id domain.ContestID) error {
	if _, err := lookup(ctx, r.registry, id); err != nil {
		return err
	}

	progress.Report(ctx, progress.Event{Phase: "analyzing"})
	return r.recomputeFindings(ctx, id)
}

func (r *Runner) recomputeFindings(ctx context.Context, id domain.ContestID) error {
	signals, err := detect.MultiStage(r.pipeline.ForContest(id)...).Run(ctx, r.submissionStore)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	findings := r.scorer.Score(signals)
	return r.findingsRepo.Put(ctx, contests.FindingsSnapshot{
		ContestID: id, Findings: findings, ComputedAt: time.Now().UTC(),
	})
}

func lookup(ctx context.Context, registry contests.ContestRegistry, id domain.ContestID) (contests.ContestRecord, error) {
	record, ok, err := registry.Get(ctx, id)
	if err != nil {
		return contests.ContestRecord{}, err
	}
	if !ok {
		return contests.ContestRecord{}, contests.ErrContestNotFound
	}
	return record, nil
}
