package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/importer"
	"scainer/internal/services/statements"
	"scainer/pkg/jobs"
	"scainer/pkg/store"
)

type countingImporter struct {
	n *atomic.Int32
}

func (countingImporter) Name() string { return "counting" }

func (c countingImporter) Import(context.Context, importer.Store) (importer.Result, error) {
	c.n.Add(1)
	return importer.Result{}, nil
}

type fakeRegistry struct {
	records []contests.ContestRecord
}

func (f *fakeRegistry) Put(context.Context, contests.ContestRecord) error { return nil }
func (f *fakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	for _, r := range f.records {
		if r.Contest.ID == id {
			return r, true, nil
		}
	}
	return contests.ContestRecord{}, false, nil
}
func (f *fakeRegistry) Delete(context.Context, domain.ContestID) error { return nil }
func (f *fakeRegistry) List(context.Context) ([]contests.ContestRecord, error) {
	return f.records, nil
}

type fakeAnalysis struct{}

func (fakeAnalysis) Put(context.Context, contests.AnalysisSnapshot) error { return nil }
func (fakeAnalysis) Get(context.Context, domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	return contests.AnalysisSnapshot{}, false, nil
}
func (fakeAnalysis) Delete(context.Context, domain.ContestID) error { return nil }

func TestImportSchedulerQueuesContests(t *testing.T) {
	var imports atomic.Int32
	importer.Register("counting", func(*yaml.Node) (importer.Importer, error) {
		return countingImporter{n: &imports}, nil
	})

	reg := &fakeRegistry{
		records: []contests.ContestRecord{
			{Contest: contests.Contest{ID: "c1"}, Source: contests.SourceSpec{Type: "counting"}},
			{Contest: contests.Contest{ID: "c2"}, Source: contests.SourceSpec{Type: "counting"}},
		},
	}
	st := store.NewMem()
	orch := analyze.NewOrchestrator(analyze.NewRegistry(detect.NewLimiter(4)))
	runner := analyze.NewRunner(reg, st, fakeAnalysis{}, orch)
	stmt := statements.NewService(reg, nil, statements.NewProblemStore(t.TempDir()))
	svc := analyze.New(jobs.NewPool(4), analyze.NewRefreshOrchestrator(runner, stmt), reg)

	sched := NewImportScheduler(reg, svc, time.Hour, "cron-user")
	sched.tick(context.Background())

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if imports.Load() >= 2 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected imports for both contests, got %d", imports.Load())
}
