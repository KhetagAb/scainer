package analyze_test

import (
	"context"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/analyze/detect/dummy"
	"scainer/pkg/store"
)

type fakeDetector struct {
	name  string
	calls *int
}

func (d fakeDetector) Name() string { return d.name }
func (d fakeDetector) AI() bool     { return false }

func (d fakeDetector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	*d.calls++
	return dummy.AlwaysProblem{}.Analyze(ctx, u)
}

func TestOrchestrator_RunsDetectorAndStoresSignals(t *testing.T) {
	ctx := context.Background()
	calls := new(int)
	orch := testOrchestrator(fakeDetector{name: "fake", calls: calls})
	st := store.NewMem()

	subs := []domain.Submission{
		{ID: "1", Contest: "c1", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c1", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	}
	if err := st.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}

	next, err := orch.Run(ctx, "c1", st, contests.AnalysisSnapshot{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if *calls == 0 {
		t.Fatal("detector was not called")
	}
	sigs := next.Signals["fake"]["problem:A:cpp"]
	if len(sigs) == 0 {
		t.Fatal("expected signals stored under scope key")
	}
	if next.ComputedAt.IsZero() {
		t.Fatal("expected ComputedAt to be set")
	}
}

var _ detect.Detector[domain.ProblemUnit] = fakeDetector{}
