package analyze_test

import (
	"context"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/analyze"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/analyze/detect/dummy"
	"scainer/internal/services/analyze/selectors"
	"scainer/internal/services/contests"
	"scainer/pkg/store"
)

func testRegistry[U domain.Unit](
	det detect.Detector[U],
	policy analyze.InvalidationPolicy[U],
	selector func(domain.ContestID) selectors.Selector[U],
) *analyze.Orchestrator {
	r := analyze.NewRegistry(detect.NewLimiter(4))
	analyze.Register(r, det, policy, selector)
	return analyze.NewOrchestrator(r)
}

type countingStandalone struct {
	calls *int
}

func (d countingStandalone) Name() string { return "standalone" }
func (d countingStandalone) AI() bool     { return false }

func (d countingStandalone) Analyze(ctx context.Context, u domain.StandaloneUnit) ([]domain.Signal, error) {
	*d.calls++
	return nil, nil
}

type countingProblem struct {
	calls *int
}

func (d countingProblem) Name() string { return "problem" }
func (d countingProblem) AI() bool     { return false }

func (d countingProblem) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	*d.calls++
	return dummy.AlwaysProblem{}.Analyze(ctx, u)
}

type countingParticipant struct {
	calls *int
}

func (d countingParticipant) Name() string { return "participant" }
func (d countingParticipant) AI() bool     { return false }

func (d countingParticipant) Analyze(ctx context.Context, u domain.ProblemParticipantUnit) ([]domain.Signal, error) {
	*d.calls++
	return nil, nil
}

func putSubs(t *testing.T, ctx context.Context, st *store.Mem, subs []domain.Submission) {
	t.Helper()
	if err := st.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}
}

func TestOrchestrator_ColdStartRunsAllUnits(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testRegistry(
		countingProblem{calls: calls},
		analyze.UnlessChanged[domain.ProblemUnit]{
			Key:    analyze.ScopeKeyProblem,
			SubIDs: analyze.SubmissionIDsFromProblemUnit,
		},
		selectors.Problems,
	)
	st := store.NewMem()
	putSubs(t, ctx, st, []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	if _, err := orch.Run(ctx, "c1", st, contests.AnalysisSnapshot{}); err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Fatalf("cold start: expected 1 call, got %d", *calls)
	}
}

func TestOrchestrator_OnceSkipsOnRepeat(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testRegistry(
		countingStandalone{calls: calls},
		analyze.Once[domain.StandaloneUnit]{Key: analyze.ScopeKeySubmission},
		selectors.Submissions,
	)
	st := store.NewMem()
	putSubs(t, ctx, st, []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a")},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b")},
	})

	prev, err := orch.Run(ctx, "c1", st, contests.AnalysisSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 2 {
		t.Fatalf("first run: expected 2 calls, got %d", *calls)
	}

	*calls = 0
	if _, err := orch.Run(ctx, "c1", st, prev); err != nil {
		t.Fatal(err)
	}
	if *calls != 0 {
		t.Fatalf("repeat run: expected 0 calls, got %d", *calls)
	}
}

func TestOrchestrator_UnlessChangedSkipsSameSet(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testRegistry(
		countingProblem{calls: calls},
		analyze.UnlessChanged[domain.ProblemUnit]{
			Key:    analyze.ScopeKeyProblem,
			SubIDs: analyze.SubmissionIDsFromProblemUnit,
		},
		selectors.Problems,
	)
	st := store.NewMem()
	subs := []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	}
	putSubs(t, ctx, st, subs)

	prev, err := orch.Run(ctx, "c1", st, contests.AnalysisSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Fatalf("first run: expected 1 call, got %d", *calls)
	}

	*calls = 0
	if _, err := orch.Run(ctx, "c1", st, prev); err != nil {
		t.Fatal(err)
	}
	if *calls != 0 {
		t.Fatalf("same set: expected 0 calls, got %d", *calls)
	}
}

func TestOrchestrator_UnlessChangedRunsOnNewSubmission(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testRegistry(
		countingProblem{calls: calls},
		analyze.UnlessChanged[domain.ProblemUnit]{
			Key:    analyze.ScopeKeyProblem,
			SubIDs: analyze.SubmissionIDsFromProblemUnit,
		},
		selectors.Problems,
	)
	st := store.NewMem()
	putSubs(t, ctx, st, []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	prev, err := orch.Run(ctx, "c1", st, contests.AnalysisSnapshot{})
	if err != nil {
		t.Fatal(err)
	}

	putSubs(t, ctx, st, []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
		{ID: "3", Contest: "c1", Problem: "A", Participant: "carol", Lang: domain.LangCPP, Source: []byte("c"), Verdict: domain.VerdictOK},
	})

	*calls = 0
	if _, err := orch.Run(ctx, "c1", st, prev); err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Fatalf("new submission: expected 1 call, got %d", *calls)
	}
}

func TestOrchestrator_ManualOnlyPreservesOnAuto(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testRegistry(
		countingParticipant{calls: calls},
		analyze.ManualOnly[domain.ProblemParticipantUnit]{Key: analyze.ScopeKeyProblemParticipant},
		selectors.OkWithLast,
	)
	st := store.NewMem()
	putSubs(t, ctx, st, []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
	})

	prev, err := orch.Run(analyze.WithManual(ctx), "c1", st, contests.AnalysisSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if *calls != 1 {
		t.Fatalf("manual run: expected 1 call, got %d", *calls)
	}

	*calls = 0
	if _, err := orch.Run(ctx, "c1", st, prev); err != nil {
		t.Fatal(err)
	}
	if *calls != 0 {
		t.Fatalf("auto run: expected 0 calls, got %d", *calls)
	}
}

var (
	_ detect.Detector[domain.StandaloneUnit]        = countingStandalone{}
	_ detect.Detector[domain.ProblemUnit]           = countingProblem{}
	_ detect.Detector[domain.ProblemParticipantUnit] = countingParticipant{}
)
