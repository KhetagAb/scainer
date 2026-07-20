package aiusage_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"scainer/internal/detect/aiusage"
	"scainer/internal/domain"
)

func TestTaskDetector(t *testing.T) {
	m := &fakeModel{reply: `{"score":0.55,"summary":"слишком чисто","evidence":[{"description":"x","submission_id":"2"}]}`}
	d := aiusage.NewTaskDetector(&aiusage.Analyzer{Model: m})
	if d.Name() != "aiusage-task" || !d.AI() {
		t.Fatalf("name/ai = %s %v", d.Name(), d.AI())
	}
	u := domain.ProblemParticipantUnit{
		Problem: "A", Participant: "alice",
		Subs: []domain.Submission{
			{ID: "1", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictWA, Source: []byte("a"), SubmittedAt: time.Unix(1, 0)},
			{ID: "2", Contest: "c", Problem: "A", Participant: "alice", Verdict: domain.VerdictOK, Source: []byte("b"), SubmittedAt: time.Unix(2, 0)},
		},
	}
	sigs, err := d.Analyze(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 || sigs[0].Score != 0.55 {
		t.Fatalf("sigs=%+v", sigs)
	}
	if _, ok := sigs[0].Subject.(domain.ParticipantProblemSubject); !ok {
		t.Fatalf("subject=%T", sigs[0].Subject)
	}
	if !strings.Contains(m.last, "Посылки") {
		t.Fatalf("prompt missing sections: %q", m.last[:min(80, len(m.last))])
	}
	if !strings.Contains(m.last, "submission_id=1") || !strings.Contains(m.last, "submission_id=2") {
		t.Fatalf("prompt missing subs")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
