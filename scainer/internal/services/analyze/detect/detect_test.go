package detect_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/selectors"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/analyze/detect/dummy"
	"scainer/pkg/progress"
	"scainer/pkg/store"
)

func TestStageStampsAIFlag(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
	})

	stage := detect.NewStage(
		selectors.ProblemSelector{Contest: "c"},
		detect.NewLimiter(4),
		dummy.AlwaysProblem{},
	)
	sigs, err := stage.Run(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) == 0 {
		t.Fatal("ожидали сигналы")
	}
	for _, s := range sigs {
		if s.AI {
			t.Fatalf("dummy не AI-детектор, got AI=true")
		}
	}
}

type countingDetector struct {
	name string
	mu   *sync.Mutex

	calls   *int
	current *int
	maxSeen *int
}

func newCountingDetector(name string) countingDetector {
	return countingDetector{
		name: name, mu: &sync.Mutex{},
		calls: new(int), current: new(int), maxSeen: new(int),
	}
}

func (d countingDetector) Name() string { return d.name }
func (d countingDetector) AI() bool     { return false }

func (d countingDetector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	d.mu.Lock()
	*d.calls++
	*d.current++
	if *d.current > *d.maxSeen {
		*d.maxSeen = *d.current
	}
	d.mu.Unlock()

	<-time.After(time.Millisecond)

	d.mu.Lock()
	*d.current--
	d.mu.Unlock()

	return []domain.Signal{{
		Detector: d.name,
		Subject:  domain.NewParticipantProblemSubject(u.Subs[0].Contest, u.Problem, u.Subs[0].Participant),
		Score:    1,
	}}, nil
}

var _ detect.Detector[domain.ProblemUnit] = countingDetector{}

func TestStageRun_RespectsConcurrencyLimit(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	var subs []domain.Submission
	for i := 0; i < 6; i++ {
		subs = append(subs, domain.Submission{
			ID: domain.SubmissionID(fmt.Sprintf("s%d", i)), Contest: "c",
			Problem: domain.ProblemID(fmt.Sprintf("P%d", i)), Participant: "solo",
			Lang: domain.LangCPP, Source: []byte("x"), Verdict: domain.VerdictOK,
		})
	}
	if err := st.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}

	const limit = 2
	det := newCountingDetector("counting")
	stage := detect.NewStage(selectors.ProblemSelector{Contest: "c"}, detect.NewLimiter(limit), det)

	sigs, err := stage.Run(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 6 {
		t.Fatalf("signals: got %d, want 6 (сигналы не должны теряться/дублироваться под параллелизмом)", len(sigs))
	}
	if *det.calls != 6 {
		t.Fatalf("calls: got %d, want 6", *det.calls)
	}
	if *det.maxSeen > limit {
		t.Fatalf("maxSeen = %d, want <= %d (SetLimit не соблюдён)", *det.maxSeen, limit)
	}
}

func TestStageRun_ProgressReportsDoneTotal(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c", Problem: "A", Participant: "bob", Lang: domain.LangCPP, Source: []byte("b"), Verdict: domain.VerdictOK},
		{ID: "3", Contest: "c", Problem: "B", Participant: "alice", Lang: domain.LangPython, Source: []byte("c"), Verdict: domain.VerdictOK},
		{ID: "4", Contest: "c", Problem: "B", Participant: "bob", Lang: domain.LangPython, Source: []byte("d"), Verdict: domain.VerdictOK},
	})

	var (
		mu   sync.Mutex
		seen []progress.Event
	)
	ctx = progress.With(ctx, func(e progress.Event) {
		mu.Lock()
		seen = append(seen, e)
		mu.Unlock()
	})

	stage := detect.NewStage(selectors.ProblemSelector{Contest: "c"}, detect.NewLimiter(4), dummy.AlwaysProblem{}, dummy.AlwaysClean{})
	if _, err := stage.Run(ctx, st); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 5 {
		t.Fatalf("events: got %d, want 5: %+v", len(seen), seen)
	}
	if seen[0].Phase != "analyzing" || seen[0].Done != 0 || seen[0].Total != 4 {
		t.Fatalf("first event = %+v, want analyzing 0/4", seen[0])
	}
	maxDone := 0
	for _, e := range seen[1:] {
		if e.Phase != "analyzing" {
			t.Fatalf("phase = %q, want analyzing", e.Phase)
		}
		if e.Total != 4 {
			t.Fatalf("total = %d, want 4", e.Total)
		}
		if e.Done > maxDone {
			maxDone = e.Done
		}
	}
	if maxDone != 4 {
		t.Fatalf("max Done = %d, want 4 (последнее событие должно быть done=total)", maxDone)
	}
}

func TestStageRun_DetectorError(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "alice", Lang: domain.LangCPP, Source: []byte("a"), Verdict: domain.VerdictOK},
	})

	stage := detect.NewStage(selectors.ProblemSelector{Contest: "c"}, detect.NewLimiter(4), failingDetector{})
	if _, err := stage.Run(ctx, st); err == nil {
		t.Fatal("ожидали ошибку от детектора")
	}
}

type failingDetector struct{}

func (failingDetector) Name() string { return "failing" }
func (failingDetector) AI() bool     { return false }
func (failingDetector) Analyze(context.Context, domain.ProblemUnit) ([]domain.Signal, error) {
	return nil, errors.New("boom")
}

var _ detect.Detector[domain.ProblemUnit] = failingDetector{}

func TestMultiStage(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "a", Lang: domain.LangCPP, Source: []byte("x"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c", Problem: "A", Participant: "b", Lang: domain.LangCPP, Source: []byte("y"), Verdict: domain.VerdictOK},
	})
	s1 := detect.NewStage(selectors.ProblemSelector{Contest: "c"}, detect.NewLimiter(2), stubDet{name: "one", score: 0.5})
	s2 := detect.NewStage(selectors.ProblemSelector{Contest: "c"}, detect.NewLimiter(2), stubDet{name: "two", score: 0.9})
	sigs, err := detect.MultiStage(s1, s2).Run(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 2 {
		t.Fatalf("len=%d", len(sigs))
	}
}

type stubDet struct {
	name  string
	score float64
}

func (d stubDet) Name() string { return d.name }
func (d stubDet) AI() bool     { return false }
func (d stubDet) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	return []domain.Signal{{
		Detector: d.name,
		Subject:  domain.NewParticipantProblemSubject("c", u.Problem, "x"),
		Score:    d.score,
	}}, nil
}
