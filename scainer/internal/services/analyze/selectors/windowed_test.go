package selectors_test

import (
	"fmt"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/selectors"
)

func at(base time.Time, sec int) time.Time {
	return base.Add(time.Duration(sec) * time.Second)
}

func ids(subs []domain.Submission) []string {
	out := make([]string, len(subs))
	for i, s := range subs {
		out[i] = string(s.ID)
	}
	return out
}

func TestWindowEndingAtLastOK(t *testing.T) {
	base := time.Unix(1000, 0)
	subs := []domain.Submission{
		{ID: "1", Verdict: domain.VerdictWA, SubmittedAt: at(base, 0)},
		{ID: "2", Verdict: domain.VerdictWA, SubmittedAt: at(base, 1)},
		{ID: "3", Verdict: domain.VerdictOK, SubmittedAt: at(base, 2)},
		{ID: "4", Verdict: domain.VerdictWA, SubmittedAt: at(base, 3)},
	}
	got := selectors.WindowEndingAtLastOK(subs, 2)
	if fmt.Sprint(ids(got)) != "[2 3]" {
		t.Fatalf("got=%v", ids(got))
	}
	got4 := selectors.WindowEndingAtLastOK(subs, 4)
	if fmt.Sprint(ids(got4)) != "[1 2 3]" {
		t.Fatalf("got=%v", ids(got4))
	}
}

func TestWindowEndingAtLastOK_NoOK(t *testing.T) {
	subs := []domain.Submission{
		{ID: "1", Verdict: domain.VerdictWA, SubmittedAt: time.Unix(1, 0)},
	}
	if got := selectors.WindowEndingAtLastOK(subs, 1); got != nil {
		t.Fatalf("got=%+v", got)
	}
}

func TestWindowEndingAtLastOK_ExactSize(t *testing.T) {
	base := time.Unix(1000, 0)
	subs := []domain.Submission{
		{ID: "1", Verdict: domain.VerdictWA, SubmittedAt: at(base, 0)},
		{ID: "2", Verdict: domain.VerdictWA, SubmittedAt: at(base, 1)},
		{ID: "3", Verdict: domain.VerdictWA, SubmittedAt: at(base, 2)},
		{ID: "4", Verdict: domain.VerdictOK, SubmittedAt: at(base, 3)},
	}
	got := selectors.WindowEndingAtLastOK(subs, 4)
	if fmt.Sprint(ids(got)) != "[1 2 3 4]" {
		t.Fatalf("got=%v", ids(got))
	}
	got1 := selectors.WindowEndingAtLastOK(subs, 1)
	if fmt.Sprint(ids(got1)) != "[4]" {
		t.Fatalf("got=%v", ids(got1))
	}
}

func TestWindowEndingAtLastOK_TwoOKs_UsesLast(t *testing.T) {
	base := time.Unix(1000, 0)
	subs := []domain.Submission{
		{ID: "ok1", Verdict: domain.VerdictOK, SubmittedAt: at(base, 0)},
		{ID: "wa", Verdict: domain.VerdictWA, SubmittedAt: at(base, 1)},
		{ID: "ok2", Verdict: domain.VerdictOK, SubmittedAt: at(base, 2)},
	}
	got := selectors.WindowEndingAtLastOK(subs, 2)
	if fmt.Sprint(ids(got)) != "[wa ok2]" {
		t.Fatalf("got=%v", ids(got))
	}
}

func TestWindowEndingAtLastOK_InvalidSize(t *testing.T) {
	subs := []domain.Submission{
		{ID: "1", Verdict: domain.VerdictOK, SubmittedAt: time.Unix(1, 0)},
	}
	if got := selectors.WindowEndingAtLastOK(subs, 0); got != nil {
		t.Fatalf("got=%v", got)
	}
}
