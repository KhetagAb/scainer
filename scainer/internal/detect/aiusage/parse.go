package aiusage

import (
	"encoding/json"
	"fmt"
	"strings"
)

type modelResponse struct {
	Score    float64         `json:"score"`
	Summary  string          `json:"summary"`
	Evidence []modelEvidence `json:"evidence"`
}

type modelEvidence struct {
	Description  string `json:"description"`
	SubmissionID string `json:"submission_id"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
}

func parseModelJSON(raw string) (modelResponse, error) {
	payload := extractJSON(raw)
	if payload == "" {
		return modelResponse{}, fmt.Errorf("нет JSON-объекта")
	}
	var out modelResponse
	dec := json.NewDecoder(strings.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		// Повтор без DisallowUnknownFields — модели часто добавляют лишние поля.
		if err2 := json.Unmarshal([]byte(payload), &out); err2 != nil {
			return modelResponse{}, err2
		}
	}
	return out, nil
}

func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	// ```json ... ``` или ``` ... ```
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			s = strings.TrimSpace(s[4:])
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < start {
		return ""
	}
	return s[start : end+1]
}
