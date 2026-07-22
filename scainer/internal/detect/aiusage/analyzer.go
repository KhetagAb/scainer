package aiusage

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/pkg/llm"
)

const rawResponseMetaLimit = 2000

type Analyzer struct {
	Model llm.IntelligenceModel
}

// Ask вызывает модель и парсит ответ ok/fail (score 0/1).
// focus — посылка, к которой крепятся spans из fail-строк from:to.
func (a *Analyzer) Ask(ctx context.Context, detectorName string, subject domain.Subject, prompt string, focus domain.SubmissionID) ([]domain.Signal, error) {
	raw, err := a.Model.Prompt(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("aiusage: %w", err)
	}
	parsed, err := parseModelReply(raw)
	if err != nil {
		return nil, fmt.Errorf("aiusage: не разобрали ответ модели: %w; body=%q", err, truncateRunes(raw, 400))
	}

	score := parsed.Score
	if score != 0 && score != 1 {
		return nil, fmt.Errorf("aiusage: score вне {0,1}: %v", score)
	}

	ev := make([]domain.Evidence, 0, len(parsed.Evidence))
	for _, item := range parsed.Evidence {
		e := domain.Evidence{
			Kind:        "ai_rationale",
			Description: item.Description,
		}
		start, end := item.StartLine, item.EndLine
		subID := item.SubmissionID
		if subID == "" {
			subID = string(focus)
		}
		if subID != "" && (start > 0 || end > 0) {
			if start <= 0 && end <= 0 {
				start, end = 1, 1
			} else if start <= 0 {
				start = end
			} else if end < start {
				end = start
			}
			e.Spans = []domain.Span{{
				Submission: domain.SubmissionID(subID),
				StartLine:  start,
				EndLine:    end,
			}}
		}
		ev = append(ev, e)
	}
	if len(ev) == 0 && parsed.Summary != "" && score > 0 {
		ev = []domain.Evidence{{Kind: "ai_rationale", Description: parsed.Summary}}
	}

	meta := map[string]any{
		"raw_response": truncateRunes(raw, rawResponseMetaLimit),
	}
	if parsed.Summary != "" {
		meta["summary"] = parsed.Summary
	}

	return []domain.Signal{{
		Detector: detectorName,
		Subject:  subject,
		Score:    score,
		Evidence: ev,
		Meta:     meta,
	}}, nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
