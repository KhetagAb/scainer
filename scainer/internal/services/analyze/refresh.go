package analyze

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/refresh"
	"scainer/internal/services/statements"
)

type refreshWorker struct {
	*Runner
}

func NewRefreshOrchestrator(runner *Runner, statementsSvc *statements.Service) *refresh.Orchestrator {
	w := refreshWorker{runner}
	return refresh.New(w, w, statementsSvc)
}

func (w refreshWorker) Import(ctx context.Context, id domain.ContestID) (refresh.ImportResult, error) {
	result, err := w.Runner.Import(ctx, id)
	if err != nil {
		return refresh.ImportResult{}, err
	}
	return refresh.ImportResult{
		ImportedCount:  result.ImportedCount,
		LastImportedAt: result.LastImportedAt,
	}, nil
}

func (w refreshWorker) ResetContestData(ctx context.Context, id domain.ContestID) error {
	return w.Runner.ResetContestData(ctx, id)
}

func (w refreshWorker) Analyze(ctx context.Context, id domain.ContestID) error {
	return w.Runner.Analyze(ctx, id)
}

func RequireImported(ctx context.Context, registry contests.ContestRegistry, id domain.ContestID) error {
	record, err := lookup(ctx, registry, id)
	if err != nil {
		return err
	}
	if record.Contest.LastImportedAt == nil {
		return ErrNotImported
	}
	return nil
}
