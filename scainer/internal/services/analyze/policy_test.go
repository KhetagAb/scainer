package analyze_test

import (
	"context"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
)

func TestEveryRun_ShouldRunAlwaysTrue(t *testing.T) {
	policy := analyze.EveryRun[domain.ProblemUnit]{
		Key: analyze.ScopeKeyProblem,
	}
	u := domain.ProblemUnit{Problem: "A", Lang: domain.LangCPP}

	tests := []struct {
		name string
		prev contests.DetectorProgress
	}{
		{"empty progress", nil},
		{"with baseline", contests.DetectorProgress{
			"problem:A:cpp": {"1", "2"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !policy.ShouldRun(context.Background(), u, tt.prev) {
				t.Fatal("expected ShouldRun=true")
			}
		})
	}
}

func TestOnce_ShouldRunAndAfterRun(t *testing.T) {
	policy := analyze.Once[domain.StandaloneUnit]{Key: analyze.ScopeKeySubmission}
	u := domain.StandaloneUnit{Sub: domain.Submission{ID: "42"}}

	if !policy.ShouldRun(context.Background(), u, nil) {
		t.Fatal("cold start: expected ShouldRun=true")
	}

	prev := contests.DetectorProgress{}
	policy.AfterRun(u, &prev)
	if policy.ShouldRun(context.Background(), u, prev) {
		t.Fatal("after run: expected ShouldRun=false")
	}
	if _, ok := prev["submission:42"]; !ok {
		t.Fatal("expected progress key after AfterRun")
	}
}

func TestUnlessChanged_ShouldRun(t *testing.T) {
	policy := analyze.UnlessChanged[domain.ProblemUnit]{
		Key:    analyze.ScopeKeyProblem,
		SubIDs: analyze.SubmissionIDsFromProblemUnit,
	}
	u := domain.ProblemUnit{
		Problem: "A",
		Lang:    domain.LangCPP,
		Subs: []domain.Submission{
			{ID: "1"},
			{ID: "2"},
		},
	}
	key := contests.ScopeKey("problem:A:cpp")

	if !policy.ShouldRun(context.Background(), u, nil) {
		t.Fatal("cold start: expected ShouldRun=true")
	}

	sameBaseline := contests.DetectorProgress{key: {"2", "1"}}
	if policy.ShouldRun(context.Background(), u, sameBaseline) {
		t.Fatal("same set: expected ShouldRun=false")
	}

	changedBaseline := contests.DetectorProgress{key: {"1"}}
	if !policy.ShouldRun(context.Background(), u, changedBaseline) {
		t.Fatal("changed set: expected ShouldRun=true")
	}
}

func TestUnlessChanged_AfterRunStoresBaseline(t *testing.T) {
	policy := analyze.UnlessChanged[domain.ProblemUnit]{
		Key:    analyze.ScopeKeyProblem,
		SubIDs: analyze.SubmissionIDsFromProblemUnit,
	}
	u := domain.ProblemUnit{
		Problem: "A",
		Lang:    domain.LangCPP,
		Subs:    []domain.Submission{{ID: "1"}, {ID: "2"}},
	}

	var prev = make(contests.DetectorProgress)
	policy.AfterRun(u, &prev)
	if !analyzeSubmissionSet(prev["problem:A:cpp"], []domain.SubmissionID{"1", "2"}) {
		t.Fatalf("unexpected baseline: %v", prev["problem:A:cpp"])
	}
}

func TestManualOnly_ShouldRun(t *testing.T) {
	policy := analyze.ManualOnly[domain.ProblemParticipantUnit]{
		Key: analyze.ScopeKeyProblemParticipant,
	}
	u := domain.ProblemParticipantUnit{Problem: "A", Participant: "alice"}

	if policy.ShouldRun(context.Background(), u, nil) {
		t.Fatal("auto: expected ShouldRun=false")
	}
	if !policy.ShouldRun(analyze.WithManual(context.Background()), u, nil) {
		t.Fatal("manual: expected ShouldRun=true")
	}
}

func analyzeSubmissionSet(got, want []domain.SubmissionID) bool {
	if len(got) != len(want) {
		return false
	}
	counts := make(map[domain.SubmissionID]int, len(want))
	for _, id := range want {
		counts[id]++
	}
	for _, id := range got {
		if counts[id] == 0 {
			return false
		}
		counts[id]--
	}
	return true
}
