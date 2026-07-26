package aiusage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"scainer/internal/services/analyze/detect/aiusage"
	"scainer/internal/domain"
)

type fakeModel struct {
	reply string
	err   error
	last  string
}

func (f *fakeModel) Prompt(ctx context.Context, prompt string) (string, error) {
	f.last = prompt
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func TestAnalyzerAsk_OK(t *testing.T) {
	m := &fakeModel{reply: "ok"}
	a := &aiusage.Analyzer{Model: m}
	subj := domain.NewParticipantProblemSubject("c", "A", "p")
	sigs, err := a.Ask(context.Background(), "aiusage-task", subj, "prompt", "s9")
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 {
		t.Fatalf("len=%d", len(sigs))
	}
	sig := sigs[0]
	if sig.Detector != "aiusage-task" || sig.Score != 0 {
		t.Fatalf("sig=%+v", sig)
	}
	if sig.Meta["summary"] != "ok" {
		t.Fatalf("meta=%v", sig.Meta)
	}
}

func TestAnalyzerAsk_FailWithSpans(t *testing.T) {
	m := &fakeModel{reply: "fail\n2:5 — tutorial comments\n10:12 - style jump"}
	a := &aiusage.Analyzer{Model: m}
	sigs, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x", "s9")
	if err != nil {
		t.Fatal(err)
	}
	if sigs[0].Score != 1 {
		t.Fatalf("score=%v", sigs[0].Score)
	}
	if len(sigs[0].Evidence) != 2 {
		t.Fatalf("evidence=%+v", sigs[0].Evidence)
	}
	sp := sigs[0].Evidence[0].Spans
	if len(sp) != 1 || sp[0].Submission != "s9" || sp[0].StartLine != 2 || sp[0].EndLine != 5 {
		t.Fatalf("spans=%+v", sp)
	}
}

func TestAnalyzerAsk_FencedOK(t *testing.T) {
	m := &fakeModel{reply: "```\nok\n```"}
	a := &aiusage.Analyzer{Model: m}
	sigs, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x", "")
	if err != nil {
		t.Fatal(err)
	}
	if sigs[0].Score != 0 {
		t.Fatalf("score=%v", sigs[0].Score)
	}
}

func TestAnalyzerAsk_BadReply(t *testing.T) {
	m := &fakeModel{reply: "not a verdict"}
	a := &aiusage.Analyzer{Model: m}
	_, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x", "")
	if err == nil || !strings.Contains(err.Error(), "не разобрали") {
		t.Fatalf("err=%v", err)
	}
}

func TestAnalyzerAsk_ModelError(t *testing.T) {
	m := &fakeModel{err: errors.New("boom")}
	a := &aiusage.Analyzer{Model: m}
	_, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x", "")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err=%v", err)
	}
}
