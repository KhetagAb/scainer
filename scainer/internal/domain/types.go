package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type (
	SubmissionID  string
	ParticipantID string
	ProblemID     string
	ContestID     string
)

type Submission struct {
	ID          SubmissionID
	Participant ParticipantID
	Problem     ProblemID
	Contest     ContestID
	Lang        Lang
	Source      []byte
	SubmittedAt time.Time
	Verdict     Verdict
	Meta        map[string]any
}

// Evidence — человекочитаемое доказательство к Signal (инвариант «Объяснимость»).
type Evidence struct {
	Kind        string `json:"kind" bson:"kind"` // напр. "jplag_match", "ai_rationale"
	Description string `json:"description" bson:"description"`
	Spans       []Span `json:"spans,omitempty" bson:"spans,omitempty"`
}

// Span — участок исходника для подсветки совпадений.
type Span struct {
	Submission SubmissionID `json:"submission" bson:"submission"`
	StartLine  int          `json:"start_line" bson:"start_line"`
	EndLine    int          `json:"end_line" bson:"end_line"`
}

// Signal — единица вывода любого детектора: подозрительность субъекта + доказательства.
// AI выставляет Stage из Detector.AI(), не сам детектор в Analyze.
type Signal struct {
	Detector string         `json:"detector" bson:"detector"`
	AI       bool           `json:"ai,omitempty" bson:"ai,omitempty"`
	Subject  Subject        `json:"subject" bson:"subject"`
	Score    float64        `json:"score" bson:"score"`
	Evidence []Evidence     `json:"evidence,omitempty" bson:"evidence,omitempty"`
	Meta     map[string]any `json:"meta,omitempty" bson:"meta,omitempty"`
}

type Finding struct {
	Subject Subject  `json:"subject" bson:"subject"`
	Score   float64  `json:"score" bson:"score"`
	Signals []Signal `json:"signals,omitempty" bson:"signals,omitempty"`
}

// BSON-драйвер не умеет конструировать интерфейс Subject сам.
func (f *Finding) UnmarshalBSON(data []byte) error {
	var raw struct {
		Subject bson.Raw `bson:"subject"`
		Score   float64  `bson:"score"`
		Signals []Signal `bson:"signals"`
	}
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}
	subj, err := unmarshalSubjectBSON(raw.Subject)
	if err != nil {
		return err
	}
	f.Subject, f.Score, f.Signals = subj, raw.Score, raw.Signals
	return nil
}

func (s *Signal) UnmarshalBSON(data []byte) error {
	var raw struct {
		Detector string         `bson:"detector"`
		AI       bool           `bson:"ai"`
		Subject  bson.Raw       `bson:"subject"`
		Score    float64        `bson:"score"`
		Evidence []Evidence     `bson:"evidence"`
		Meta     map[string]any `bson:"meta"`
	}
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}
	subj, err := unmarshalSubjectBSON(raw.Subject)
	if err != nil {
		return err
	}
	*s = Signal{
		Detector: raw.Detector,
		AI:       raw.AI,
		Subject:  subj,
		Score:    raw.Score,
		Evidence: raw.Evidence,
		Meta:     raw.Meta,
	}
	return nil
}
