package analyze_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/selectors"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/importer"
	"scainer/internal/services/scoring"
	"scainer/pkg/jobs"
	"scainer/pkg/store"
)

func testOrchestrator(dets ...detect.Detector[domain.ProblemUnit]) *analyze.Orchestrator {
	r := analyze.NewRegistry(detect.NewLimiter(4))
	for _, det := range dets {
		analyze.Register(r, det, analyze.EveryRun[domain.ProblemUnit]{Key: analyze.ScopeKeyProblem}, selectors.Problems)
	}
	return analyze.NewOrchestrator(r)
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
	jobID, err := svc.SyncManual(ctx, id)
	if err != nil {
		return err
	}
	return waitForJob(t, svc, jobID)
}

type stubImporter struct{}

func (stubImporter) Name() string { return "stub" }

func (stubImporter) Import(context.Context, importer.Store) (importer.Result, error) {
	return importer.Result{}, nil
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

type fakeAnalysisRepository struct {
	byID map[domain.ContestID]contests.AnalysisSnapshot
}

func newFakeAnalysisRepository() *fakeAnalysisRepository {
	return &fakeAnalysisRepository{byID: make(map[domain.ContestID]contests.AnalysisSnapshot)}
}

func (f *fakeAnalysisRepository) Put(_ context.Context, snap contests.AnalysisSnapshot) error {
	f.byID[snap.ContestID] = snap
	return nil
}

func (f *fakeAnalysisRepository) Get(_ context.Context, id domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	snap, ok := f.byID[id]
	return snap, ok, nil
}

func (f *fakeAnalysisRepository) Delete(_ context.Context, id domain.ContestID) error {
	delete(f.byID, id)
	return nil
}

type countingDetector struct {
	calls *int
}

func (d countingDetector) Name() string { return "counting" }
func (d countingDetector) AI() bool     { return false }

func pairSignals(u domain.ProblemUnit, detector string, score float64) []domain.Signal {
	var out []domain.Signal
	for i := 0; i < len(u.Subs); i++ {
		for j := i + 1; j < len(u.Subs); j++ {
			out = append(out, domain.Signal{
				Detector: detector,
				Subject: domain.NewPairSubject(
					u.Subs[i].Contest,
					u.Problem,
					u.Subs[i].Participant,
					u.Subs[j].Participant,
				),
				Score: score,
			})
		}
	}
	return out
}

func (d countingDetector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	*d.calls++
	return pairSignals(u, "counting", 1.0), nil
}

func registerWithSubs(t *testing.T, subs []domain.Submission) (*contests.ContestReader, *analyze.Service, *int) {
	t.Helper()
	ctx := context.Background()
	calls := new(int)
	st := store.NewMem()
	reg := newFakeRegistry()
	repo := newFakeAnalysisRepository()
	orch := testOrchestrator(countingDetector{calls: calls})
	scorer := scoring.NewWeighted()

	svc := contests.NewService(reg, repo, "stub")
	reader := contests.NewContestReader(reg, st, repo, scorer)
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, repo, orch))

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

func TestSyncManual_AsyncJobFlow(t *testing.T) {
	ctx := context.Background()
	reader, analyzeSvc, _ := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	jobID, err := analyzeSvc.SyncManual(ctx, "contest01")
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

func TestSyncManual_UnknownContest(t *testing.T) {
	ctx := context.Background()
	analyzeSvc := analyze.New(
		jobs.NewPool(4),
		analyze.NewRunner(newFakeRegistry(), store.NewMem(), newFakeAnalysisRepository(), testOrchestrator()),
	)

	if _, err := analyzeSvc.SyncManual(ctx, "missing"); !errors.Is(err, contests.ErrContestNotFound) {
		t.Fatalf("got %v want ErrContestNotFound", err)
	}
}

func TestAnalyze_WithoutImport(t *testing.T) {
	ctx := context.Background()
	_, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
	})

	jobID, err := analyzeSvc.Analyze(ctx, "contest01")
	if err == nil || !errors.Is(err, analyze.ErrNotImported) {
		t.Fatalf("Analyze without import: got %v want ErrNotImported", err)
	}
	if jobID != "" {
		t.Fatalf("expected empty jobID, got %q", jobID)
	}
	if *calls != 0 {
		t.Fatalf("Analyze without import must not call detectors, got %d", *calls)
	}
}

func TestAnalyze_WithImport(t *testing.T) {
	ctx := context.Background()
	reader, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	importJobID, err := analyzeSvc.Import(ctx, "contest01")
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if err := waitForJob(t, analyzeSvc, importJobID); err != nil {
		t.Fatalf("import job: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("Import не должен вызывать детекторы, got %d calls", *calls)
	}

	analyzeJobID, err := analyzeSvc.Analyze(ctx, "contest01")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if err := waitForJob(t, analyzeSvc, analyzeJobID); err != nil {
		t.Fatalf("analyze job: %v", err)
	}
	if *calls == 0 {
		t.Fatal("ожидали вызов детектора после Analyze")
	}

	findings, _, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки после Analyze")
	}
}

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
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, testOrchestrator()))

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

	secondJobID, err := analyzeSvc.SyncManual(ctx, "contest01")
	if err != nil {
		t.Fatalf("second SyncManual: %v", err)
	}
	if secondJobID != jobID {
		t.Fatalf("got job %q want same job %q", secondJobID, jobID)
	}

	close(release)
	if err := waitForJob(t, analyzeSvc, jobID); err != nil {
		t.Fatalf("job: %v", err)
	}
}
