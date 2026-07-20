package detect

import (
	"testing"
	"time"

	"scainer/internal/domain"
)

func TestLatestPerParticipant(t *testing.T) {
	t0 := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)

	subs := []domain.Submission{
		{ID: "ejudge:1:1", Participant: "alice", SubmittedAt: t0},
		{ID: "ejudge:1:3", Participant: "bob", SubmittedAt: t0},
		{ID: "ejudge:1:2", Participant: "alice", SubmittedAt: t1},
		{ID: "ejudge:1:4", Participant: "carol", SubmittedAt: t1},
	}
	got := LatestPerParticipant(subs)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ID != "ejudge:1:2" || got[0].Participant != "alice" {
		t.Fatalf("alice = %+v", got[0])
	}
	if got[1].ID != "ejudge:1:3" || got[1].Participant != "bob" {
		t.Fatalf("bob = %+v", got[1])
	}
	if got[2].ID != "ejudge:1:4" {
		t.Fatalf("carol = %+v", got[2])
	}
}

func TestLatestPerParticipant_TieBreakByID(t *testing.T) {
	ts := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	subs := []domain.Submission{
		{ID: "ejudge:1:01", Participant: "alice", SubmittedAt: ts},
		{ID: "ejudge:1:10", Participant: "alice", SubmittedAt: ts},
		{ID: "ejudge:1:02", Participant: "alice", SubmittedAt: ts},
	}
	got := LatestPerParticipant(subs)
	if len(got) != 1 || got[0].ID != "ejudge:1:10" {
		t.Fatalf("got = %+v", got)
	}
}

func TestLatestPerParticipant_Empty(t *testing.T) {
	if got := LatestPerParticipant(nil); len(got) != 0 {
		t.Fatalf("got = %+v", got)
	}
}

func TestLatestSuccessfulPerParticipant(t *testing.T) {
	t0 := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	t2 := t1.Add(time.Hour)

	subs := []domain.Submission{
		{ID: "1", Participant: "alice", SubmittedAt: t0, Verdict: domain.VerdictOK},
		{ID: "2", Participant: "alice", SubmittedAt: t1, Verdict: domain.VerdictWA}, // новее, но не OK
		{ID: "3", Participant: "alice", SubmittedAt: t2, Verdict: domain.VerdictOK}, // последняя OK
		{ID: "4", Participant: "bob", SubmittedAt: t0, Verdict: domain.VerdictML},   // нет OK — выбыл
		{ID: "5", Participant: "carol", SubmittedAt: t0, Verdict: domain.VerdictOK},
	}
	got := LatestSuccessfulPerParticipant(subs)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].ID != "3" || got[0].Participant != "alice" {
		t.Fatalf("alice = %+v", got[0])
	}
	if got[1].ID != "5" || got[1].Participant != "carol" {
		t.Fatalf("carol = %+v", got[1])
	}
}
