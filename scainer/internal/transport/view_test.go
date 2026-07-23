package transport

import (
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/detect/nightsubmit"
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

func TestBuildData_SubmissionSubjectReferenced(t *testing.T) {
	subID := domain.SubmissionID("ejudge:50601:1")
	subj := domain.NewSubmissionSubject("50601", subID)
	findings := []domain.Finding{{
		Subject: subj,
		Score:   1,
		Signals: []domain.Signal{{
			Detector: "night-submit",
			Subject:  subj,
			Score:    1,
			Evidence: []domain.Evidence{{
				Kind:        "night_submit",
				Description: nightsubmit.EvidenceDescription(mskTime(2, 14)),
			}},
		}},
	}}
	subs := map[domain.SubmissionID]domain.Submission{
		subID: {
			ID: subID, Contest: "50601", Participant: "alice", Problem: "A",
			Source: []byte("int main() {}"),
		},
	}

	data := BuildData(findings, subs)
	if len(data.Findings) != 1 {
		t.Fatalf("findings: %d", len(data.Findings))
	}
	fv := data.Findings[0]
	if len(fv.Subject.Participants) != 1 || fv.Subject.Participants[0] != "alice" {
		t.Fatalf("participants = %#v", fv.Subject.Participants)
	}
	if fv.Subject.Problem != "A" {
		t.Fatalf("problem = %q", fv.Subject.Problem)
	}
	if _, ok := data.Submissions[string(subID)]; !ok {
		t.Fatal("submission not in Submissions map")
	}
}

func mskTime(hour, min int) time.Time {
	return time.Date(2026, 7, 23, hour, min, 0, 0, time.FixedZone("MSK", 3*60*60))
}
