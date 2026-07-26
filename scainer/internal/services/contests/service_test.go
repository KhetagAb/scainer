package contests_test

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
	"scainer/internal/services/analyze/detect/dummy"
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
	jobID, err := svc.ImportThenAnalyze(ctx, id)
	if err != nil {
		return err
	}
	return waitForJob(t, svc, jobID)
}

type stubImporter struct {
	subs []domain.Submission
}

func (stubImporter) Name() string { return "stub" }

func (s stubImporter) Import(context.Context, importer.Store) (importer.Result, error) {
	return importer.Result{Submissions: s.subs}, nil
}

var _ importer.Importer = stubImporter{}

func init() {
	importer.Register("stub", func(node *yaml.Node) (importer.Importer, error) {
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

var _ contests.ContestRegistry = (*fakeRegistry)(nil)

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

var _ contests.AnalysisRepository = (*fakeAnalysisRepository)(nil)

type countingDetector struct {
	calls *int
}

func (d countingDetector) Name() string { return "counting" }

func (d countingDetector) AI() bool { return false }

func (d countingDetector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	*d.calls++
	return dummy.AlwaysProblem{}.Analyze(ctx, u)
}

var _ detect.Detector[domain.ProblemUnit] = countingDetector{}

// registerWithSubs — контест с заранее положенными сабмитами (register + analyze + reader).
func registerWithSubs(t *testing.T, subs []domain.Submission) (*contests.Service, *contests.ContestReader, *analyze.Service, *int) {
	t.Helper()
	ctx := context.Background()
	calls := new(int)
	st := store.NewMem()
	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	det := countingDetector{calls: calls}
	orch := testOrchestrator(det)

	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := st.Put(ctx, subs); err != nil {
		t.Fatalf("Put: %v", err)
	}

	return svc, reader, analyzeSvc, calls
}

func TestRegisterSuccess(t *testing.T) {
	ctx := context.Background()
	svc := contests.NewService(newFakeRegistry(), newFakeAnalysisRepository(), "stub")

	decl := contests.Registration{
		ID:               "contest01",
		ParallelID:       "par1",
		ExcludedProblems: []domain.ProblemID{"Z"},
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	info, err := svc.Register(ctx, decl)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if info.ID != "contest01" {
		t.Fatalf("ID: got %v want contest01", info.ID)
	}
	if info.ParallelID != "par1" {
		t.Fatalf("ParallelID: got %v want par1", info.ParallelID)
	}
	if len(info.ExcludedProblems) != 1 || info.ExcludedProblems[0] != "Z" {
		t.Fatalf("ExcludedProblems: got %v want [Z]", info.ExcludedProblems)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	ctx := context.Background()
	svc := contests.NewService(newFakeRegistry(), newFakeAnalysisRepository(), "stub")

	decl := contests.Registration{
		ID: "contest01",
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("First Register: %v", err)
	}

	if _, err := svc.Register(ctx, decl); !errors.Is(err, contests.ErrDuplicateContest) {
		t.Fatalf("Second Register: got %v want ErrDuplicateContest", err)
	}
}

func TestRegistryIsSourceOfTruthAcrossServiceRestart(t *testing.T) {
	ctx := context.Background()
	registry := newFakeRegistry()
	fs := newFakeAnalysisRepository()

	svc1 := contests.NewService(registry, fs, "stub")

	decl := contests.Registration{
		ID:               "contest01",
		ParallelID:       "par1",
		ExcludedProblems: []domain.ProblemID{"A", "B"},
		Source:           &contests.SourceSpec{Type: "stub"},
	}
	if _, err := svc1.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := svc1.SetParallel(ctx, "contest01", "par2"); err != nil {
		t.Fatalf("SetParallel: %v", err)
	}

	svc2 := contests.NewService(registry, newFakeAnalysisRepository(), "stub")
	reader2 := contests.NewContestReader(registry, store.NewMem(), newFakeAnalysisRepository(), scoring.NewWeighted())

	list, err := reader2.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ожидали 1 контест, получили %d", len(list))
	}
	got := list[0]
	if got.ParallelID != "par2" {
		t.Fatalf("параллель не пережила рестарт: %+v", got)
	}
	if len(got.ExcludedProblems) != 2 {
		t.Fatalf("ExcludedProblems не пережили рестарт: %+v", got.ExcludedProblems)
	}

	if _, err := svc2.Register(ctx, decl); !errors.Is(err, contests.ErrDuplicateContest) {
		t.Fatalf("Register: got %v want ErrDuplicateContest", err)
	}
}

func TestSetParallel(t *testing.T) {
	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs, scoring.NewWeighted())

	decl := contests.Registration{
		ID: "contest01",
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.SetParallel(ctx, "contest01", "par2"); err != nil {
		t.Fatalf("SetParallel: %v", err)
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list) != 1 || list[0].ParallelID != "par2" {
		t.Fatalf("List: got %+v", list[0])
	}
}

func TestRemoveContest(t *testing.T) {
	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs, scoring.NewWeighted())

	decl := contests.Registration{
		ID: "contest01",
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.RemoveContest(ctx, "contest01"); err != nil {
		t.Fatalf("RemoveContest: %v", err)
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list) != 0 {
		t.Fatalf("List: expected 0 contests, got %d", len(list))
	}
}

func TestSetExcludedProblems(t *testing.T) {
	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs, scoring.NewWeighted())

	decl := contests.Registration{
		ID: "contest01",
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}

	excluded := []domain.ProblemID{"A", "B"}
	if err := svc.SetExcludedProblems(ctx, "contest01", excluded); err != nil {
		t.Fatalf("SetExcludedProblems: %v", err)
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list[0].ExcludedProblems) != 2 {
		t.Fatalf("ExcludedProblems: got %v want 2 items", len(list[0].ExcludedProblems))
	}
}

func TestSetExcludedProblemsDoesNotTriggerRecompute(t *testing.T) {
	ctx := context.Background()
	svc, reader, imports, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, imports, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	afterImport := *calls

	if err := svc.SetExcludedProblems(ctx, "contest01", []domain.ProblemID{"A"}); err != nil {
		t.Fatalf("SetExcludedProblems: %v", err)
	}

	if *calls != afterImport {
		t.Fatalf("SetExcludedProblems не должен триггерить пересчёт: было %d вызовов, стало %d", afterImport, *calls)
	}

	findings, _, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("findings должны остаться от последнего Import (не пересчитаны с учётом нового исключения)")
	}
}
