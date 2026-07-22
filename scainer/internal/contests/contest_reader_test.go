package contests_test

import (
	"context"
	"testing"

	"scainer/internal/contests"
	"scainer/internal/domain"
	"scainer/internal/store"
)

func TestListSortsByContestID(t *testing.T) {
	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs)

	for _, id := range []domain.ContestID{"10", "2", "9"} {
		if _, err := svc.Register(ctx, contests.Registration{
			ID:     id,
			Source: &contests.SourceSpec{Type: "stub"},
		}); err != nil {
			t.Fatalf("Register %s: %v", id, err)
		}
	}

	list, err := reader.List(ctx)
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

func TestListStatsZeroBeforeImport(t *testing.T) {
	ctx := context.Background()
	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs)

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	list, err := reader.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len: got %d", len(list))
	}
	got := list[0]
	if got.Statistic.SubmissionCount != 0 || got.Statistic.ProblemCount != 0 {
		t.Fatalf("counts: got subs=%d problems=%d", got.Statistic.SubmissionCount, got.Statistic.ProblemCount)
	}
}

func TestProblemsPendingCount(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	reg := newFakeRegistry()
	findingsStore := newFakeFindingsStore()
	svc := contests.NewService(reg, findingsStore, "stub")
	reader := contests.NewContestReader(reg, st, findingsStore)

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictPR},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
		{ID: "3", Contest: "contest01", Problem: "A", Participant: "carol", Lang: domain.LangCPP, Source: []byte("c"), Verdict: domain.VerdictPR},
		{ID: "4", Contest: "contest01", Problem: "B", Participant: "alice", Lang: domain.LangCPP, Source: []byte("d"), Verdict: domain.VerdictWA},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	problems, err := reader.Problems(ctx, "contest01")
	if err != nil {
		t.Fatalf("Problems: %v", err)
	}
	byID := map[domain.ProblemID]contests.ProblemInfo{}
	for _, p := range problems {
		byID[p.ID] = p
	}
	a := byID["A"]
	if a.SubmissionCount != 3 || a.PendingCount != 2 {
		t.Fatalf("A: subs=%d pending=%d, want 3/2", a.SubmissionCount, a.PendingCount)
	}
	b := byID["B"]
	if b.SubmissionCount != 1 || b.PendingCount != 0 {
		t.Fatalf("B: subs=%d pending=%d, want 1/0", b.SubmissionCount, b.PendingCount)
	}
}

func TestListEnrichStats(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	reg := newFakeRegistry()
	findingsStore := newFakeFindingsStore()
	svc := contests.NewService(reg, findingsStore, "stub")
	reader := contests.NewContestReader(reg, st, findingsStore)

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

	list, err := reader.List(ctx)
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
}

func TestGetFindingsDoesNotRecompute(t *testing.T) {
	ctx := context.Background()
	_, reader, analyzeSvc, calls := registerWithSubs(t, []domain.Submission{
		{ID: "1", Contest: "contest01", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "contest01", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if err := submitAndWait(t, analyzeSvc, ctx, "contest01"); err != nil {
		t.Fatalf("Import: %v", err)
	}
	afterImport := *calls
	if afterImport == 0 {
		t.Fatal("детектор не вызывался при Import")
	}

	for i := 0; i < 3; i++ {
		findings, _, err := reader.GetFindings(ctx, "contest01")
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
	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, store.NewMem(), fs)

	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	findings, subs, err := reader.GetFindings(ctx, "contest01")
	if err != nil {
		t.Fatalf("GetFindings: %v", err)
	}
	if findings != nil || subs != nil {
		t.Fatalf("ожидали (nil, nil) до Import, получили (%v, %v)", findings, subs)
	}
}
