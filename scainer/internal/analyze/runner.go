package analyze

import (
	"context"
	"fmt"
	"time"

	"scainer/internal/contests"
	"scainer/internal/detect"
	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/internal/scoring"
)

type Result struct {
	ImportedCount  int
	LastImportedAt time.Time
}

type Runner struct {
	registry        contests.ContestRegistry
	submissionStore contests.SubmissionStore
	findingsStore   contests.FindingsStore
	scorer          scoring.Scorer
	pipeline        Pipeline
}

func NewRunner(
	registry contests.ContestRegistry,
	submissionStore contests.SubmissionStore,
	findingsStore contests.FindingsStore,
	scorer scoring.Scorer,
	pipeline Pipeline,
) *Runner {
	return &Runner{
		registry:        registry,
		submissionStore: submissionStore,
		findingsStore:   findingsStore,
		scorer:          scorer,
		pipeline:        pipeline,
	}
}

func (r *Runner) Run(ctx context.Context, id domain.ContestID) (Result, error) {
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

	if err := r.recomputeFindings(ctx, id); err != nil {
		return Result{}, fmt.Errorf("recompute findings: %w", err)
	}
	return Result{ImportedCount: len(imported.Submissions), LastImportedAt: now}, nil
}

func (r *Runner) recomputeFindings(ctx context.Context, id domain.ContestID) error {
	signals, err := detect.MultiStage(r.pipeline.ForContest(id)...).Run(ctx, r.submissionStore)
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	findings := r.scorer.Score(signals)
	return r.findingsStore.Put(ctx, contests.FindingsSnapshot{
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
