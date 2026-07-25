package analyze_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/importer"
	"scainer/internal/services/scoring"
	"scainer/pkg/jobs"
	"scainer/pkg/store"
)

func TestImport_SkipsAnalyze(t *testing.T) {
	ctx := context.Background()
	reader, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
	})

	jobID, err := analyzeSvc.Import(ctx, "contest01")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if err := waitForJob(t, analyzeSvc, jobID); err != nil {
		t.Fatalf("job: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("Import не должен вызывать детекторы, got %d calls", *calls)
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].LastImportedAt == nil {
		t.Fatalf("ожидали LastImportedAt после Import, got %+v", list)
	}
}

type blockingImporter struct {
	release chan struct{}
}

func (blockingImporter) Name() string { return "blocking" }

func (b blockingImporter) Import(ctx context.Context, _ importer.Store) (importer.Result, error) {
	select {
	case <-b.release:
		return importer.Result{}, nil
	case <-ctx.Done():
		return importer.Result{}, ctx.Err()
	}
}

func TestJobConflict(t *testing.T) {
	release := make(chan struct{})
	importer.Register("blocking", func(*yaml.Node) (importer.Importer, error) {
		return blockingImporter{release: release}, nil
	})

	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeFindingsRepository()
	st := store.NewMem()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), testPipeline()))

	svc := contests.NewService(reg, fs, "blocking")
	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "blocking"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	jobID, err := analyzeSvc.Import(ctx, "contest01")
	if err != nil {
		t.Fatalf("first Import: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st, ok := analyzeSvc.JobStatus(jobID)
		if ok && st.Status == jobs.StatusRunning {
			break
		}
		time.Sleep(time.Millisecond)
	}

	_, err = analyzeSvc.ImportThenAnalyze(ctx, "contest01")
	if !errors.Is(err, analyze.ErrJobRunning) {
		t.Fatalf("got %v want ErrJobRunning", err)
	}

	close(release)
	if err := waitForJob(t, analyzeSvc, jobID); err != nil {
		t.Fatalf("job: %v", err)
	}
}
