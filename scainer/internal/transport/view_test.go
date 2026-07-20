package transport

import (
	"testing"

	"github.com/lksh/scainer/internal/domain"
)

func TestBuildReportData_ReferencedOnlyAndAIFlag(t *testing.T) {
	pair := domain.NewPairSubject("contest01", "A", "alice", "bob")
	findings := []domain.Finding{{
		Subject: pair,
		Score:   1,
		Signals: []domain.Signal{
			{
				Detector: "jplag",
				AI:       false,
				Subject:  pair,
				Score:    1,
				Evidence: []domain.Evidence{{
					Kind:        "jplag_match",
					Description: "match",
					Spans: []domain.Span{
						{Submission: "contest01/A/alice", StartLine: 1, EndLine: 2},
						{Submission: "contest01/A/bob", StartLine: 1, EndLine: 2},
					},
				}},
			},
			{
				Detector: "llm-check",
				AI:       true,
				Score:    0.5,
			},
		},
	}}
	subs := map[domain.SubmissionID]domain.Submission{
		"contest01/A/alice": {ID: "contest01/A/alice", Participant: "alice", Problem: "A", Source: []byte("a\nb")},
		"contest01/A/bob":   {ID: "contest01/A/bob", Participant: "bob", Problem: "A", Source: []byte("a\nb")},
		"unused":            {ID: "unused", Participant: "carol", Problem: "B", Source: []byte("x")},
	}

	data := BuildData(findings, subs)
	if len(data.Findings) != 1 {
		t.Fatalf("findings: %d", len(data.Findings))
	}
	if !data.Findings[0].AI {
		t.Fatal("ожидали AI=true на карточке (есть AI-сигнал)")
	}
	if data.Findings[0].Signals[0].AI {
		t.Fatal("jplag не AI")
	}
	if !data.Findings[0].Signals[1].AI {
		t.Fatal("llm-check должен быть AI")
	}
	if _, ok := data.Submissions["unused"]; ok {
		t.Fatal("unused submission не должна попадать в Submissions")
	}
	if len(data.Submissions) != 2 {
		t.Fatalf("submissions: %d", len(data.Submissions))
	}
}
