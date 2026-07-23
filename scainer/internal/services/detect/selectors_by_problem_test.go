package detect_test

import (
	"context"
	"testing"
	"time"

	"scainer/internal/services/detect"
	"scainer/internal/domain"
	"scainer/pkg/store"
)

func TestOkWithLastSelector(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	base := time.Unix(1000, 0)
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictWA, SubmittedAt: base},
		{ID: "2", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictWA, SubmittedAt: base.Add(time.Second)},
		{ID: "3", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictOK, SubmittedAt: base.Add(2 * time.Second)},
		{ID: "4", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictWA, SubmittedAt: base.Add(3 * time.Second)}, // after OK
		{ID: "5", Contest: "c", Problem: "B", Participant: "alice", Verdict: domain.VerdictOK, SubmittedAt: base},
		{ID: "6", Contest: "c", Problem: "A", Participant: "bob", Verdict: domain.VerdictWA, SubmittedAt: base}, // нет OK
		{ID: "7", Contest: "c", Problem: "X", Participant: "alice", Verdict: domain.VerdictWA, SubmittedAt: base},
		{ID: "8", Contest: "c", Problem: "X", Participant: "alice", Verdict: domain.VerdictOK, SubmittedAt: base.Add(time.Second)},
	})

	units, err := detect.OkWithLastSelector{Contest: "c"}.Select(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 3 {
		t.Fatalf("units=%d %+v", len(units), units)
	}
	var aliceA, aliceB *domain.ProblemParticipantUnit
	for i := range units {
		switch {
		case units[i].Problem == "A" && units[i].Participant == "alice":
			aliceA = &units[i]
		case units[i].Problem == "B" && units[i].Participant == "alice":
			aliceB = &units[i]
		}
	}
	if aliceA == nil || len(aliceA.Subs) != 3 || aliceA.Subs[0].ID != "1" || aliceA.Subs[2].ID != "3" {
		t.Fatalf("aliceA=%+v", aliceA)
	}
	if aliceB == nil || len(aliceB.Subs) != 1 || aliceB.Subs[0].ID != "5" {
		t.Fatalf("aliceB=%+v", aliceB)
	}
}

func TestOkWithLastSelector_WindowCap(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	base := time.Unix(1000, 0)
	var subs []domain.Submission
	for i := 0; i < 6; i++ {
		v := domain.VerdictWA
		if i == 5 {
			v = domain.VerdictOK
		}
		subs = append(subs, domain.Submission{
			ID: domain.SubmissionID(string(rune('1' + i))), Contest: "c", Problem: "A",
			Participant: "p", Verdict: v, SubmittedAt: base.Add(time.Duration(i) * time.Second),
		})
	}
	_ = st.Put(ctx, subs)

	units, err := detect.OkWithLastSelector{Contest: "c"}.Select(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || len(units[0].Subs) != detect.OkWithLastMin {
		t.Fatalf("units=%+v", units)
	}
}

func TestMultiStage(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "A", Participant: "a", Lang: domain.LangCPP, Source: []byte("x"), Verdict: domain.VerdictOK},
		{ID: "2", Contest: "c", Problem: "A", Participant: "b", Lang: domain.LangCPP, Source: []byte("y"), Verdict: domain.VerdictOK},
	})
	s1 := detect.NewStage(detect.ProblemSelector{Contest: "c"}, detect.NewLimiter(2), stubDet{name: "one", score: 0.5})
	s2 := detect.NewStage(detect.ProblemSelector{Contest: "c"}, detect.NewLimiter(2), stubDet{name: "two", score: 0.9})
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
