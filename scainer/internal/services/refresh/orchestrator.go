package refresh

import (
	"context"
	"errors"
	"log"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/statements"
)

type ImportResult struct {
	ImportedCount  int
	LastImportedAt time.Time
}

type Runner interface {
	Import(ctx context.Context, id domain.ContestID) (ImportResult, error)
	ResetContestData(ctx context.Context, id domain.ContestID) error
}

type Analyzer interface {
	Analyze(ctx context.Context, id domain.ContestID) error
}

type Orchestrator struct {
	runner     Runner
	analyzer   Analyzer
	statements *statements.Service
}

func New(runner Runner, analyzer Analyzer, statementsSvc *statements.Service) *Orchestrator {
	return &Orchestrator{runner: runner, analyzer: analyzer, statements: statementsSvc}
}

func (o *Orchestrator) Import(ctx context.Context, id domain.ContestID) (ImportResult, error) {
	result, err := o.runner.Import(ctx, id)
	if err != nil {
		return result, err
	}
	o.refreshStatements(ctx, id)
	return result, nil
}

func (o *Orchestrator) Reset(ctx context.Context, id domain.ContestID) error {
	if err := o.runner.ResetContestData(ctx, id); err != nil {
		return err
	}
	return o.statements.InvalidateContestStatements(id)
}

func (o *Orchestrator) Sync(ctx context.Context, id domain.ContestID) error {
	if _, err := o.Import(ctx, id); err != nil {
		return err
	}
	return o.analyzer.Analyze(ctx, id)
}

func (o *Orchestrator) Resync(ctx context.Context, id domain.ContestID) error {
	if err := o.Reset(ctx, id); err != nil {
		return err
	}
	if _, err := o.Import(ctx, id); err != nil {
		return err
	}
	return o.analyzer.Analyze(ctx, id)
}

func (o *Orchestrator) Analyze(ctx context.Context, id domain.ContestID) error {
	return o.analyzer.Analyze(ctx, id)
}

func (o *Orchestrator) refreshStatements(ctx context.Context, id domain.ContestID) {
	if err := o.statements.RefreshContestStatements(ctx, id); err != nil {
		if errors.Is(err, statements.ErrNotAvailable) {
			return
		}
		log.Printf("refresh: contest=%s statements: %v", id, err)
	}
}
