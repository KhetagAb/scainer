package nightsubmit_test

import (
	"context"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/detect/nightsubmit"
)

func TestDetector_NightSubmission(t *testing.T) {
	d := nightsubmit.NewDetector()
	sub := domain.Submission{
		ID:          "ejudge:1:42",
		Contest:     "50601",
		Participant: "alice",
		Problem:     "A",
		SubmittedAt: mskTime(2, 14),
	}
	sigs, err := d.Analyze(context.Background(), domain.StandaloneUnit{Sub: sub})
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 {
		t.Fatalf("signals: got %d, want 1", len(sigs))
	}
	if sigs[0].Detector != nightsubmit.DetectorName {
		t.Fatalf("detector = %q", sigs[0].Detector)
	}
	if sigs[0].Score != 1.0 {
		t.Fatalf("score = %v", sigs[0].Score)
	}
	subj, ok := sigs[0].Subject.(domain.SubmissionSubject)
	if !ok || subj.Submission != sub.ID {
		t.Fatalf("subject = %#v", sigs[0].Subject)
	}
	if len(sigs[0].Evidence) != 1 || sigs[0].Evidence[0].Kind != "night_submit" {
		t.Fatalf("evidence = %#v", sigs[0].Evidence)
	}
	if len(sigs[0].Evidence[0].Spans) != 1 || sigs[0].Evidence[0].Spans[0].Submission != sub.ID {
		t.Fatalf("spans = %#v", sigs[0].Evidence[0].Spans)
	}
	wantDesc := nightsubmit.EvidenceDescription(sub.SubmittedAt)
	if sigs[0].Evidence[0].Description != wantDesc {
		t.Fatalf("description = %q, want %q", sigs[0].Evidence[0].Description, wantDesc)
	}
}

func TestDetector_DaySubmission(t *testing.T) {
	d := nightsubmit.NewDetector()
	sub := domain.Submission{
		ID:          "ejudge:1:43",
		Contest:     "50601",
		SubmittedAt: mskTime(14, 0),
	}
	sigs, err := d.Analyze(context.Background(), domain.StandaloneUnit{Sub: sub})
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 0 {
		t.Fatalf("expected no signals, got %#v", sigs)
	}
}

func TestDetector_ZeroSubmittedAt(t *testing.T) {
	d := nightsubmit.NewDetector()
	sub := domain.Submission{ID: "x", Contest: "c"}
	sigs, err := d.Analyze(context.Background(), domain.StandaloneUnit{Sub: sub})
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 0 {
		t.Fatalf("expected no signals, got %#v", sigs)
	}
}
