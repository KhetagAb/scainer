package aiusage_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"scainer/internal/services/analyze/detect/aiusage"
	"scainer/internal/domain"
)

func TestTaskDetector(t *testing.T) {
	m := &fakeModel{reply: "fail\n1:2 — too clean"}
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
	if len(sigs) != 1 || sigs[0].Score != 1 {
		t.Fatalf("sigs=%+v", sigs)
	}
	if _, ok := sigs[0].Subject.(domain.ParticipantProblemSubject); !ok {
		t.Fatalf("subject=%T", sigs[0].Subject)
	}
	if !strings.Contains(m.last, "Current submission") || !strings.Contains(m.last, "Previous submission") {
		t.Fatalf("prompt missing sections: %q", m.last[:min(120, len(m.last))])
	}
	if !strings.Contains(m.last, "submission_id=1") || !strings.Contains(m.last, "submission_id=2") {
		t.Fatalf("prompt missing subs")
	}
	if len(sigs[0].Evidence) == 0 || len(sigs[0].Evidence[0].Spans) == 0 || sigs[0].Evidence[0].Spans[0].Submission != "2" {
		t.Fatalf("focus spans=%+v", sigs[0].Evidence)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
