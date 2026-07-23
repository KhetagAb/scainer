package aiusage

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type modelResponse struct {
	Score    float64
	Summary  string
	Evidence []modelEvidence
}

type modelEvidence struct {
	Description  string
	SubmissionID string
	StartLine    int
	EndLine      int
}

var failSpanLine = regexp.MustCompile(`(?i)^\s*(\d+)\s*:\s*(\d+)\s*[-–—]+\s*(.*\S)?\s*$`)

// parseModelReply — формат ok / fail (+ строки from:to — comment).
// ok → score 0, fail → score 1.
func parseModelReply(raw string) (modelResponse, error) {
	s := strings.TrimSpace(stripMarkdownFence(raw))
	if s == "" {
		return modelResponse{}, fmt.Errorf("пустой ответ")
	}
	lines := splitLines(s)
	if len(lines) == 0 {
		return modelResponse{}, fmt.Errorf("пустой ответ")
	}
	first := strings.ToLower(strings.TrimSpace(lines[0]))
	// Иногда модель пишет "fail:" / "ok." — берём первое слово.
	firstWord := first
	if i := strings.IndexAny(first, " \t:.,;"); i >= 0 {
		firstWord = first[:i]
	}
	switch firstWord {
	case "ok":
		return modelResponse{Score: 0, Summary: "ok"}, nil
	case "fail":
		out := modelResponse{Score: 1, Summary: "fail"}
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			m := failSpanLine.FindStringSubmatch(line)
			if m == nil {
				out.Evidence = append(out.Evidence, modelEvidence{Description: line})
				continue
			}
			start, _ := strconv.Atoi(m[1])
			end, _ := strconv.Atoi(m[2])
			comment := strings.TrimSpace(m[3])
			if comment == "" {
				comment = line
			}
			if end < start {
				end = start
			}
			out.Evidence = append(out.Evidence, modelEvidence{
				Description: comment,
				StartLine:   start,
				EndLine:     end,
			})
		}
		return out, nil
	default:
		return modelResponse{}, fmt.Errorf("ожидали ok или fail, got %q", truncateForErr(lines[0], 80))
	}
}

func stripMarkdownFence(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		first := strings.TrimSpace(s[:i])
		rest := s[i+1:]
		// Язык ограждения (json/text) vs уже контент (ok/fail).
		if first == "" || (first != "ok" && first != "fail" && !strings.Contains(first, ":")) {
			low := strings.ToLower(first)
			if first == "" || low == "json" || low == "text" || low == "markdown" {
				s = rest
			}
		}
	}
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

func truncateForErr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// parseModelJSON — старый JSON-контракт {score,summary,evidence}; оставлен для отката.
func parseModelJSON(raw string) (modelResponse, error) {
	payload := extractJSON(raw)
	if payload == "" {
		return modelResponse{}, fmt.Errorf("нет JSON-объекта")
	}
	var wire struct {
		Score    float64 `json:"score"`
		Summary  string  `json:"summary"`
		Evidence []struct {
			Description  string `json:"description"`
			SubmissionID string `json:"submission_id"`
			StartLine    int    `json:"start_line"`
			EndLine      int    `json:"end_line"`
		} `json:"evidence"`
	}
	dec := json.NewDecoder(strings.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&wire); err != nil {
		if err2 := json.Unmarshal([]byte(payload), &wire); err2 != nil {
			return modelResponse{}, err2
		}
	}
	out := modelResponse{Score: wire.Score, Summary: wire.Summary}
	for _, item := range wire.Evidence {
		out.Evidence = append(out.Evidence, modelEvidence{
			Description:  item.Description,
			SubmissionID: item.SubmissionID,
			StartLine:    item.StartLine,
			EndLine:      item.EndLine,
		})
	}
	return out, nil
}

func extractJSON(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
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
