package contests_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/lksh/scainer/internal/contests"
	"github.com/lksh/scainer/internal/detect"
	"github.com/lksh/scainer/internal/detect/dummy"
	"github.com/lksh/scainer/internal/domain"
	"github.com/lksh/scainer/internal/importer"
	"github.com/lksh/scainer/internal/jobs"
	"github.com/lksh/scainer/internal/scoring"
	"github.com/lksh/scainer/internal/store"
)

func newRuntimeConfig() contests.RuntimeConfig {
	return contests.RuntimeConfig{
		Pool:           jobs.NewPool(4),
		AnalyzeLimiter: detect.NewLimiter(4),
	}
}

func waitForJob(t *testing.T, svc *contests.Service, jobID string) error {
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

func submitAndWait(t *testing.T, svc *contests.Service, ctx context.Context, id domain.ContestID) error {
	t.Helper()
	jobID, err := svc.SubmitImport(ctx, id)
	if err != nil {
		return err
	}
	return waitForJob(t, svc, jobID)
}

type stubImporter struct {
	subs []domain.Submission
}

func (stubImporter) Name() string { return "stub" }

func (s stubImporter) Import(context.Context, store.Store) ([]domain.Submission, error) {
	return s.subs, nil
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

var _ contests.FindingsStore = (*fakeFindingsStore)(nil)

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

func TestRegisterSuccess(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

	decl := contests.Registration{
		ID:               "contest01",
		ParallelID:       "par1",
		ParallelName:     "Параллель 1",
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
	if info.ParallelName != "Параллель 1" {
		t.Fatalf("ParallelName: got %v want Параллель 1", info.ParallelName)
	}
	if len(info.ExcludedProblems) != 1 || info.ExcludedProblems[0] != "Z" {
		t.Fatalf("ExcludedProblems: got %v want [Z]", info.ExcludedProblems)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

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

func TestSetParallel(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

	decl := contests.Registration{
		ID: "contest01",
		Source: &contests.SourceSpec{
			Type: "stub",
		},
	}

	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := svc.SetParallel(ctx, "contest01", "par2", "Параллель 2"); err != nil {
		t.Fatalf("SetParallel: %v", err)
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list) != 1 || list[0].ParallelID != "par2" || list[0].ParallelName != "Параллель 2" {
		t.Fatalf("List: got %+v", list[0])
	}
}

func TestRemoveContest(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

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

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list) != 0 {
		t.Fatalf("List: expected 0 contests, got %d", len(list))
	}
}

func TestSetExcludedProblems(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

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

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list[0].ExcludedProblems) != 2 {
		t.Fatalf("ExcludedProblems: got %v want 2 items", len(list[0].ExcludedProblems))
	}
}

func TestRegistryIsSourceOfTruthAcrossServiceRestart(t *testing.T) {
	ctx := context.Background()
	registry := newFakeRegistry()

	svc1 := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", registry, newFakeFindingsStore(), newRuntimeConfig())

	decl := contests.Registration{
		ID:               "contest01",
		ParallelID:       "par1",
		ParallelName:     "Параллель 1",
		ExcludedProblems: []domain.ProblemID{"A", "B"},
		Source:           &contests.SourceSpec{Type: "stub"},
	}
	if _, err := svc1.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := svc1.SetParallel(ctx, "contest01", "par2", "Параллель 2"); err != nil {
		t.Fatalf("SetParallel: %v", err)
	}

	svc2 := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", registry, newFakeFindingsStore(), newRuntimeConfig())

	list, err := svc2.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ожидали 1 контест, получили %d", len(list))
	}
	got := list[0]
	if got.ParallelID != "par2" || got.ParallelName != "Параллель 2" {
		t.Fatalf("параллель не пережила рестарт: %+v", got)
	}
	if len(got.ExcludedProblems) != 2 {
		t.Fatalf("ExcludedProblems не пережили рестарт: %+v", got.ExcludedProblems)
	}

	if _, err := svc2.Register(ctx, decl); !errors.Is(err, contests.ErrDuplicateContest) {
		t.Fatalf("Register: got %v want ErrDuplicateContest", err)
	}
}

func TestListSortsByContestID(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

	for _, id := range []domain.ContestID{"10", "2", "9"} {
		if _, err := svc.Register(ctx, contests.Registration{
			ID:     id,
			Source: &contests.SourceSpec{Type: "stub"},
		}); err != nil {
			t.Fatalf("Register %s: %v", id, err)
		}
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []domain.ContestID{"2", "9", "10"}
	if len(list) != len(want) {
		t.Fatalf("len: got %d want %d", len(list), len(want))
	}
	for i, id := range want {
		if list[i].ID != id {
			t.Fatalf("List[%d]: got %s want %s", i, list[i].ID, id)
		}
	}
}

func registerWithSubs(t *testing.T, subs []domain.Submission) (*contests.Service, *int) {
	t.Helper()
	ctx := context.Background()
	calls := new(int)
	st := store.NewMem()
	svc := contests.New(st, scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig(), countingDetector{calls: calls})

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// stubImporter вернёт nil; Put заранее — Store.Put(nil) no-op.
	if err := st.Put(ctx, subs); err != nil {
		t.Fatalf("Put: %v", err)
	}

	return svc, calls
}

func TestImportComputesFindings(t *testing.T) {
	ctx := context.Background()
	svc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	findings, _, err := svc.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки от counting-детектора, получили 0 — детектор не подключён к стадии")
	}
	if *calls == 0 {
		t.Fatal("детектор не вызывался")
	}
}

func TestGetFindingsDoesNotRecompute(t *testing.T) {
	ctx := context.Background()
	svc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	afterImport := *calls
	if afterImport == 0 {
		t.Fatal("детектор не вызывался при Import")
	}

	for i := 0; i < 3; i++ {
		findings, _, err := svc.GetFindings(ctx, "contest01")
		if err != nil {
			t.Fatalf("GetFindings: %v", err)
		}
		if len(findings) == 0 {
			t.Fatal("GetFindings вернул пусто после Import")
		}
	}

	if *calls != afterImport {
		t.Fatalf("GetFindings пересчитал: было %d вызовов детектора, стало %d", afterImport, *calls)
	}
}

func TestGetFindingsBeforeImport_EmptyNotError(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	findings, subs, err := svc.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if findings != nil || subs != nil {
		t.Fatalf("ожидали (nil, nil) до Import, получили (%v, %v)", findings, subs)
	}
}

func TestImportRecomputesOnReimport(t *testing.T) {
	ctx := context.Background()
	svc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("first Import: %v", err)
	}
	afterFirst := *calls
	if afterFirst == 0 {
		t.Fatal("детектор не вызывался при первом Import")
	}

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("second Import: %v", err)
	}
	if *calls <= afterFirst {
		t.Fatalf("повторный Import не пересчитал: было %d, стало %d", afterFirst, *calls)
	}
}

func TestSetExcludedProblemsDoesNotTriggerRecompute(t *testing.T) {
	ctx := context.Background()
	svc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	afterImport := *calls

	if err := svc.SetExcludedProblems(ctx, "contest01", []domain.ProblemID{"A"}); err != nil {
		t.Fatalf("SetExcludedProblems: %v", err)
	}

	if *calls != afterImport {
		t.Fatalf("SetExcludedProblems не должен триггерить пересчёт: было %d вызовов, стало %d", afterImport, *calls)
	}

	findings, _, err := svc.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("findings должны остаться от последнего Import (не пересчитаны с учётом нового исключения)")
	}
}

func TestListStatsZeroBeforeImport(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())
	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len: got %d", len(list))
	}
	got := list[0]
	if got.Statistic.SubmissionCount != 0 || got.Statistic.ProblemCount != 0 || got.Statistic.FindingsCount != 0 {
		t.Fatalf("counts: got subs=%d problems=%d findings=%d", got.Statistic.SubmissionCount, got.Statistic.ProblemCount, got.Statistic.FindingsCount)
	}
	if got.Statistic.WeightedSuspicionPercent != nil {
		t.Fatalf("percent: want nil, got %v", *got.Statistic.WeightedSuspicionPercent)
	}
}

func TestListEnrichStats(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	findingsStore := newFakeFindingsStore()
	svc := contests.New(st, scoring.NewWeighted(), "stub", newFakeRegistry(), findingsStore, newRuntimeConfig())

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
		{ID: "3", Contest: "contest01", Problem: "B", Participant: "alice", Lang: domain.LangCPP, Source: []byte("c"), Verdict: domain.VerdictOK},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := findingsStore.Put(ctx, contests.FindingsSnapshot{
		ContestID: "contest01",
		Findings: []domain.Finding{
			{Score: 0.5},
			{Score: 1.0},
		},
	}); err != nil {
		t.Fatalf("findings Put: %v", err)
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got := list[0]
	if got.Statistic.SubmissionCount != 3 {
		t.Fatalf("SubmissionCount: got %d want 3", got.Statistic.SubmissionCount)
	}
	if got.Statistic.ProblemCount != 2 {
		t.Fatalf("ProblemCount: got %d want 2", got.Statistic.ProblemCount)
	}
	if got.Statistic.FindingsCount != 2 {
		t.Fatalf("FindingsCount: got %d want 2", got.Statistic.FindingsCount)
	}
	// 100 * (0.5+1.0) / 3 = 50
	if got.Statistic.WeightedSuspicionPercent == nil || *got.Statistic.WeightedSuspicionPercent != 50 {
		t.Fatalf("WeightedSuspicionPercent: got %v want 50", got.Statistic.WeightedSuspicionPercent)
	}
}

func TestAnalyzeUsesRegisteredDetectors(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	svc := contests.New(st, scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig(), dummy.AlwaysProblem{})

	decl := contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}
	if _, err := svc.Register(ctx, decl); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if err := submitAndWait(t, svc, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	findings, _, err := svc.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки от AlwaysProblem-детектора, получили 0 — детектор не подключён к стадии")
	}
}

func TestSubmitImport_AsyncJobFlow(t *testing.T) {
	ctx := context.Background()
	svc, _ := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	jobID, err := svc.SubmitImport(ctx, "contest01")
	if err != nil {
		t.Fatalf("SubmitImport: %v", err)
	}
	if jobID == "" {
		t.Fatal("ожидали непустой jobID")
	}

	if err := waitForJob(t, svc, jobID); err != nil {
		t.Fatalf("job: %v", err)
	}
	st, ok := svc.JobStatus(jobID)
	if !ok {
		t.Fatal("JobStatus: job не найден")
	}
	if st.Status != jobs.StatusSucceeded {
		t.Fatalf("status: got %v want succeeded", st.Status)
	}

	findings, _, err := svc.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("ожидали находки после завершения job'а")
	}
}

func TestSubmitImport_UnknownContest(t *testing.T) {
	ctx := context.Background()
	svc := contests.New(store.NewMem(), scoring.NewWeighted(), "stub", newFakeRegistry(), newFakeFindingsStore(), newRuntimeConfig())

	if _, err := svc.SubmitImport(ctx, "missing"); !errors.Is(err, contests.ErrContestNotFound) {
		t.Fatalf("got %v want ErrContestNotFound", err)
	}
}
