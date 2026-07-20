package analyze_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/analyze"
	"scainer/internal/contests"
	"scainer/internal/detect"
	"scainer/internal/detect/dummy"
	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/internal/jobs"
	"scainer/internal/scoring"
	"scainer/internal/store"
)

func testPipeline(dets ...detect.Detector[domain.ProblemUnit]) detect.Pipeline {
	limiter := detect.NewLimiter(4)
	if len(dets) == 0 {
		return detect.Compose()
	}
	return detect.Compose(func(id domain.ContestID) detect.Stage {
		return detect.NewStage(detect.ProblemSelector{Contest: id}, limiter, dets...)
	})
}

func waitForJob(t *testing.T, svc *analyze.Service, jobID string) error {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		st, ok := svc.JobStatus(jobID)
		if !ok {
			t.Fatalf("job %s не найден", jobID)
		}
		switch st.Status {
		case jobs.StatusSucceeded:
			return nil
		case jobs.StatusFailed:
			return errors.New(st.Err)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job не завершился до таймаута")
	return nil
}

func submitAndWait(t *testing.T, svc *analyze.Service, ctx context.Context, id domain.ContestID) error {
	t.Helper()
	jobID, err := svc.Submit(ctx, id)
	if err != nil {
		return err
	}
	return waitForJob(t, svc, jobID)
}

type stubImporter struct{}

func (stubImporter) Name() string { return "stub" }

func (stubImporter) Import(context.Context, importer.Store) ([]domain.Submission, error) {
	return nil, nil
}

func init() {
	importer.Register("stub", func(*yaml.Node) (importer.Importer, error) {
		return stubImporter{}, nil
	})
}

type fakeRegistry struct {
	byID map[domain.ContestID]contests.ContestRecord
}

func newFakeRegistry() *fakeRegistry {
	return &fakeRegistry{byID: make(map[domain.ContestID]contests.ContestRecord)}
}

func (r *fakeRegistry) Put(_ context.Context, rec contests.ContestRecord) error {
	r.byID[rec.Contest.ID] = rec
	return nil
}

func (r *fakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	rec, ok := r.byID[id]
	return rec, ok, nil
}

func (r *fakeRegistry) Delete(_ context.Context, id domain.ContestID) error {
	delete(r.byID, id)
	return nil
}

func (r *fakeRegistry) List(context.Context) ([]contests.ContestRecord, error) {
	out := make([]contests.ContestRecord, 0, len(r.byID))
	for _, rec := range r.byID {
		out = append(out, rec)
	}
	return out, nil
}

type fakeFindingsStore struct {
	byID map[domain.ContestID]contests.FindingsSnapshot
}

func newFakeFindingsStore() *fakeFindingsStore {
	return &fakeFindingsStore{byID: make(map[domain.ContestID]contests.FindingsSnapshot)}
}

func (f *fakeFindingsStore) Put(_ context.Context, snap contests.FindingsSnapshot) error {
	f.byID[snap.ContestID] = snap
	return nil
}

func (f *fakeFindingsStore) Get(_ context.Context, id domain.ContestID) (contests.FindingsSnapshot, bool, error) {
	snap, ok := f.byID[id]
	return snap, ok, nil
}

func (f *fakeFindingsStore) Delete(_ context.Context, id domain.ContestID) error {
	delete(f.byID, id)
	return nil
}

type countingDetector struct {
	calls *int
}

func (d countingDetector) Name() string { return "counting" }
func (d countingDetector) AI() bool     { return false }

func (d countingDetector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	*d.calls++
	return dummy.AlwaysProblem{}.Analyze(ctx, u)
}

func registerWithSubs(t *testing.T, subs []domain.Submission) (*contests.ContestReader, *analyze.Service, *int) {
	t.Helper()
	ctx := context.Background()
	calls := new(int)
	st := store.NewMem()
	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	pipeline := testPipeline(countingDetector{calls: calls})

	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := st.Put(ctx, subs); err != nil {
		t.Fatalf("Put: %v", err)
	}
	return reader, analyzeSvc, calls
}

func TestComputesFindings(t *testing.T) {
	ctx := context.Background()
	reader, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, analyzeSvc, ctx, "contest01"); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	findings, _, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки от counting-детектора, получили 0")
	}
	if *calls == 0 {
		t.Fatal("детектор не вызывался")
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].LastImportedAt == nil {
		t.Fatalf("ожидали LastImportedAt после Submit, got %+v", list)
	}
}

func TestRecomputesOnResubmit(t *testing.T) {
	ctx := context.Background()
	_, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, analyzeSvc, ctx, "contest01"); err != nil {
		t.Fatalf("first Submit: %v", err)
	}
	afterFirst := *calls
	if afterFirst == 0 {
		t.Fatal("детектор не вызывался при первом Submit")
	}

	if err := submitAndWait(t, analyzeSvc, ctx, "contest01"); err != nil {
		t.Fatalf("second Submit: %v", err)
	}
	if *calls <= afterFirst {
		t.Fatalf("повторный Submit не пересчитал: было %d, стало %d", afterFirst, *calls)
	}
}

func TestUsesPipelineDetectors(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	pipeline := testPipeline(dummy.AlwaysProblem{})

	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if err := submitAndWait(t, analyzeSvc, ctx, "contest01"); err != nil {
		t.Fatalf("Submit: %v", err)
	}
	findings, _, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки от AlwaysProblem-детектора")
	}
}

func TestSubmit_AsyncJobFlow(t *testing.T) {
	ctx := context.Background()
	reader, analyzeSvc, _ := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	jobID, err := analyzeSvc.Submit(ctx, "contest01")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if jobID == "" {
		t.Fatal("ожидали непустой jobID")
	}
	if err := waitForJob(t, analyzeSvc, jobID); err != nil {
		t.Fatalf("job: %v", err)
	}
	st, ok := analyzeSvc.JobStatus(jobID)
	if !ok {
		t.Fatal("JobStatus: job не найден")
	}
	if st.Status != jobs.StatusSucceeded {
		t.Fatalf("status: got %v want succeeded", st.Status)
	}

	findings, _, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки после завершения job'а")
	}
}

func TestSubmit_UnknownContest(t *testing.T) {
	ctx := context.Background()
	analyzeSvc := analyze.New(
		jobs.NewPool(4),
		analyze.NewRunner(newFakeRegistry(), store.NewMem(), newFakeFindingsStore(), scoring.NewWeighted(), testPipeline()),
	)

	if _, err := analyzeSvc.Submit(ctx, "missing"); !errors.Is(err, contests.ErrContestNotFound) {
		t.Fatalf("got %v want ErrContestNotFound", err)
	}
}
