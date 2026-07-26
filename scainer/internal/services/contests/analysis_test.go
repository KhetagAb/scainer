package contests_test

import (
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/scoring"
)

func TestFlattenSignals_DeterministicOrder(t *testing.T) {
	signals := contests.Signals{
		"z-det": {
			"b": {{Detector: "z-det", Score: 2}},
			"a": {{Detector: "z-det", Score: 1}},
		},
		"a-det": {
			"y": {{Detector: "a-det", Score: 4}},
			"x": {{Detector: "a-det", Score: 3}},
		},
	}

	flat := contests.FlattenSignals(signals)
	if len(flat) != 4 {
		t.Fatalf("len: got %d want 4", len(flat))
	}
	want := []float64{3, 4, 1, 2}
	for i, score := range want {
		if flat[i].Score != score {
			t.Fatalf("flat[%d].Score: got %v want %v", i, flat[i].Score, score)
		}
	}
}

func TestFlattenSignals_Empty(t *testing.T) {
	if got := contests.FlattenSignals(nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestAnalysisSnapshot_Findings_DeriveOnRead(t *testing.T) {
	contest := domain.ContestID("c1")
	subj := domain.NewPairSubject(contest, "A", "alice", "bob")
	snap := contests.AnalysisSnapshot{
		ContestID: contest,
		Signals: contests.Signals{
			"jplag": {
				"problem:A:cpp": {
					{Detector: "jplag", Subject: subj, Score: 0.9},
				},
			},
		},
	}

	findings := snap.Findings(scoring.NewWeighted())
	if len(findings) != 1 {
		t.Fatalf("len: got %d want 1", len(findings))
	}
	if findings[0].Score != 0.9 {
		t.Fatalf("score: got %v want 0.9", findings[0].Score)
	}
	if domain.SubjectKey(findings[0].Subject) != domain.SubjectKey(subj) {
		t.Fatalf("subject: got %#v want %#v", findings[0].Subject, subj)
	}
}

func TestAnalysisSnapshot_Findings_EmptySnapshot(t *testing.T) {
	if got := (*contests.AnalysisSnapshot)(nil).Findings(scoring.NewWeighted()); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestAnalysisSnapshot_Clone_PreservesOnceProgress(t *testing.T) {
	snap := contests.AnalysisSnapshot{
		Progress: contests.Progress{
			"night-submit": {
				"submission:1": nil,
			},
		},
	}

	cloned := snap.Clone()
	prog := cloned.Progress["night-submit"]
	if _, ok := prog["submission:1"]; !ok {
		t.Fatal("expected Once progress key to survive Clone")
	}
	if prog["submission:1"] != nil {
		t.Fatalf("expected nil baseline slice, got %v", prog["submission:1"])
	}
}
