package aiusage_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"scainer/internal/detect/aiusage"
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
	m := &fakeModel{reply: `{
		"score": 0.7,
		"summary": "Скачок стиля",
		"evidence": [{"description":"другие комментарии","submission_id":"s9","start_line":2,"end_line":5}]
	}`}
	a := &aiusage.Analyzer{Model: m}
	subj := domain.NewParticipantProblemSubject("c", "A", "p")
	sigs, err := a.Ask(context.Background(), "aiusage-task", subj, "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 {
		t.Fatalf("len=%d", len(sigs))
	}
	sig := sigs[0]
	if sig.Detector != "aiusage-task" || sig.Score != 0.7 {
		t.Fatalf("sig=%+v", sig)
	}
	if len(sig.Evidence) != 1 || sig.Evidence[0].Kind != "ai_rationale" {
		t.Fatalf("evidence=%+v", sig.Evidence)
	}
	if len(sig.Evidence[0].Spans) != 1 || sig.Evidence[0].Spans[0].Submission != "s9" {
		t.Fatalf("spans=%+v", sig.Evidence[0].Spans)
	}
	if sig.Meta["summary"] != "Скачок стиля" {
		t.Fatalf("meta=%v", sig.Meta)
	}
}

func TestAnalyzerAsk_FencedJSON(t *testing.T) {
	m := &fakeModel{reply: "```json\n{\"score\":0.1,\"summary\":\"ok\",\"evidence\":[]}\n```"}
	a := &aiusage.Analyzer{Model: m}
	sigs, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x")
	if err != nil {
		t.Fatal(err)
	}
	if sigs[0].Score != 0.1 {
		t.Fatalf("score=%v", sigs[0].Score)
	}
}

func TestAnalyzerAsk_BadJSON(t *testing.T) {
	m := &fakeModel{reply: "not json at all"}
	a := &aiusage.Analyzer{Model: m}
	_, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x")
	if err == nil || !strings.Contains(err.Error(), "битый JSON") {
		t.Fatalf("err=%v", err)
	}
}

func TestAnalyzerAsk_ModelError(t *testing.T) {
	m := &fakeModel{err: errors.New("boom")}
	a := &aiusage.Analyzer{Model: m}
	_, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err=%v", err)
	}
}

func TestAnalyzerAsk_ScoreOutOfRange(t *testing.T) {
	m := &fakeModel{reply: `{"score":1.5,"summary":"x","evidence":[]}`}
	a := &aiusage.Analyzer{Model: m}
	_, err := a.Ask(context.Background(), "d", domain.NewParticipantSubject("p"), "x")
	if err == nil || !strings.Contains(err.Error(), "score") {
		t.Fatalf("err=%v", err)
	}
}
