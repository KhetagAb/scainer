package analyze

import (
	"context"
	"errors"
	"fmt"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/ejudge"
	"scainer/internal/services/importer"
	"scainer/pkg/progress"
)

var ErrNotImported = errors.New("contest has not been imported")

func NewRunner(
	registry contests.ContestRegistry,
	submissionStore contests.SubmissionStore,
	analysisRepo contests.AnalysisRepository,
	orchestrator *Orchestrator,
) *Runner {
	return &Runner{
		registry:        registry,
		submissionStore: submissionStore,
		analysisRepo:    analysisRepo,
		orchestrator:    orchestrator,
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

func (r *Runner) ResetContestData(ctx context.Context, id domain.ContestID) error {
	record, err := lookup(ctx, r.registry, id)
	if err != nil {
		return err
	}

	if err := r.submissionStore.DeleteByContest(ctx, id); err != nil {
		return fmt.Errorf("delete submissions: %w", err)
	}
	if err := r.analysisRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete analysis: %w", err)
	}
	if record.Source.Type == "ejudge" {
		var cfg ejudge.ImportConfig
		if record.Source.Config.Kind != 0 {
			if err := record.Source.Config.Decode(&cfg); err != nil {
				return fmt.Errorf("decode ejudge config: %w", err)
			}
		}
		if cfg.ContestID > 0 {
			if err := r.submissionStore.DeleteCursor(ctx, ejudge.ImportCursorKey(cfg.ContestID)); err != nil {
				return fmt.Errorf("delete ejudge cursor: %w", err)
			}
		}
	}

	contest := record.Contest
	contest.LastImportedAt = nil
	if err := r.registry.Put(ctx, contests.ContestRecord{Contest: contest, Source: record.Source}); err != nil {
		return fmt.Errorf("clear lastImportedAt: %w", err)
	}
	return nil
}

func (r *Runner) Analyze(ctx context.Context, id domain.ContestID) error {
	if err := r.requireImported(ctx, id); err != nil {
		return err
	}

	progress.Report(ctx, progress.Event{Phase: "analyzing"})
	return r.recomputeAnalysis(ctx, id)
}

func (r *Runner) requireImported(ctx context.Context, id domain.ContestID) error {
	record, err := lookup(ctx, r.registry, id)
	if err != nil {
		return err
	}
	if record.Contest.LastImportedAt == nil {
		return ErrNotImported
	}
	return nil
}

func (r *Runner) recomputeAnalysis(ctx context.Context, id domain.ContestID) error {
	prev, ok, err := r.analysisRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		prev = contests.AnalysisSnapshot{ContestID: id}
	}

	next, err := r.orchestrator.Run(ctx, id, r.submissionStore, prev)
	if putErr := r.analysisRepo.Put(ctx, next); putErr != nil {
		return putErr
	}
	if err != nil {
		return fmt.Errorf("analyze: %w", err)
	}
	return nil
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
