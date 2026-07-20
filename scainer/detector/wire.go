package detector

import "encoding/json"

type Request struct {
	UnitKind string          `json:"unit_kind"` // "standalone" | "pair" | "problem" | "problem_participant" | "participant"
	Unit     json.RawMessage `json:"unit"`
}

type Response struct {
	Signals []SignalDTO `json:"signals"`
	Error   string      `json:"error,omitempty"`
}

type SignalDTO struct {
	Detector    string          `json:"detector"`
	AI          bool            `json:"ai,omitempty"`
	SubjectKind string          `json:"subject_kind"` // "submission" | "pair" | "participant_problem" | "participant"
	Subject     json.RawMessage `json:"subject"`
	Score       float64         `json:"score"`
	Evidence    []EvidenceDTO   `json:"evidence,omitempty"`
	Meta        map[string]any  `json:"meta,omitempty"`
}

type EvidenceDTO struct {
	Kind        string    `json:"kind"`
	Description string    `json:"description"`
	Spans       []SpanDTO `json:"spans,omitempty"`
}

type SpanDTO struct {
	Submission string `json:"submission"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}
