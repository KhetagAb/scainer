package transport

import (
	"fmt"
	"strings"
	"time"

	"github.com/lksh/scainer/internal/domain"
)

type Data struct {
	GeneratedNote string                    `json:"generated_note"`
	Findings      []FindingView             `json:"findings"`
	Submissions   map[string]SubmissionView `json:"submissions"`
}

type FindingView struct {
	Key     string       `json:"key"`
	Subject SubjectView  `json:"subject"`
	Score   float64      `json:"score"`
	AI      bool         `json:"ai,omitempty"`
	Signals []SignalView `json:"signals"`
}

type SubjectView struct {
	Kind         string   `json:"kind"`
	Contest      string   `json:"contest,omitempty"`
	ContestName  string   `json:"contest_name,omitempty"`
	Participants []string `json:"participants,omitempty"`
	Problem      string   `json:"problem,omitempty"`
	ProblemName  string   `json:"problem_name,omitempty"`
	Submission   string   `json:"submission,omitempty"`
}

type SignalView struct {
	Detector string         `json:"detector"`
	AI       bool           `json:"ai,omitempty"`
	Score    float64        `json:"score"`
	Evidence []EvidenceView `json:"evidence"`
}

type EvidenceView struct {
	Kind        string     `json:"kind"`
	Description string     `json:"description"`
	Spans       []SpanView `json:"spans,omitempty"`
}

type SpanView struct {
	Submission string `json:"submission"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}

type SubmissionView struct {
	Contest     string   `json:"contest,omitempty"`
	ContestName string   `json:"contest_name,omitempty"`
	Participant string   `json:"participant"`
	Problem     string   `json:"problem"`
	ProblemName string   `json:"problem_name,omitempty"`
	Lang        string   `json:"lang"`
	SubmittedAt string   `json:"submitted_at"`
	Verdict     string   `json:"verdict"`
	Source      []string `json:"source"`
}

func BuildData(findings []domain.Finding, subs map[domain.SubmissionID]domain.Submission) Data {
	labels := indexDisplayNames(subs)
	referenced := make(map[domain.SubmissionID]bool)
	out := Data{
		GeneratedNote: formatSignalCount(len(findings)),
		Findings:      make([]FindingView, 0, len(findings)),
		Submissions:   make(map[string]SubmissionView),
	}

	for _, f := range findings {
		fv := FindingView{
			Key:     domain.SubjectKey(f.Subject),
			Subject: toSubjectView(f.Subject, labels),
			Score:   f.Score,
			Signals: make([]SignalView, 0, len(f.Signals)),
		}
		for _, sig := range f.Signals {
			if sig.AI {
				fv.AI = true
			}
			sv := SignalView{
				Detector: sig.Detector,
				AI:       sig.AI,
				Score:    sig.Score,
				Evidence: make([]EvidenceView, 0, len(sig.Evidence)),
			}
			for _, ev := range sig.Evidence {
				evv := EvidenceView{
					Kind:        ev.Kind,
					Description: ev.Description,
				}
				for _, sp := range ev.Spans {
					referenced[sp.Submission] = true
					evv.Spans = append(evv.Spans, SpanView{
						Submission: string(sp.Submission),
						StartLine:  sp.StartLine,
						EndLine:    sp.EndLine,
					})
				}
				sv.Evidence = append(sv.Evidence, evv)
			}
			fv.Signals = append(fv.Signals, sv)
		}
		out.Findings = append(out.Findings, fv)
	}

	for id := range referenced {
		s, ok := subs[id]
		if !ok {
			continue
		}
		out.Submissions[string(id)] = toSubmissionView(s)
	}
	return out
}

type displayNames struct {
	contest map[string]string
	problem map[string]string // contestID\x00problemID
}

func indexDisplayNames(subs map[domain.SubmissionID]domain.Submission) displayNames {
	out := displayNames{
		contest: make(map[string]string),
		problem: make(map[string]string),
	}
	for _, s := range subs {
		c := string(s.Contest)
		if name := metaString(s.Meta, "contest_name"); name != "" {
			out.contest[c] = name
		}
		if name := metaString(s.Meta, "problem_name"); name != "" && s.Problem != "" {
			out.problem[c+"\x00"+string(s.Problem)] = name
		}
	}
	return out
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	v, ok := meta[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toSubjectView(s domain.Subject, labels displayNames) SubjectView {
	v := SubjectView{Kind: string(s.Kind())}
	switch x := s.(type) {
	case domain.SubmissionSubject:
		v.Contest = string(x.Contest)
		v.Submission = string(x.Submission)
	case domain.PairSubject:
		v.Contest = string(x.Contest)
		v.Problem = string(x.Problem)
		v.Participants = []string{string(x.A), string(x.B)}
	case domain.ParticipantProblemSubject:
		v.Contest = string(x.Contest)
		v.Problem = string(x.Problem)
		v.Participants = []string{string(x.Participant)}
	case domain.ParticipantSubject:
		v.Participants = []string{string(x.Participant)}
	default:
		panic(fmt.Sprintf("transport: unknown Subject %T", s))
	}
	if v.Contest != "" {
		v.ContestName = labels.contest[v.Contest]
	}
	if v.Contest != "" && v.Problem != "" {
		v.ProblemName = labels.problem[v.Contest+"\x00"+v.Problem]
	}
	return v
}

func toSubmissionView(s domain.Submission) SubmissionView {
	src := string(s.Source)
	lines := strings.Split(src, "\n")
	submitted := ""
	if !s.SubmittedAt.IsZero() {
		submitted = s.SubmittedAt.UTC().Format(time.RFC3339)
	}
	return SubmissionView{
		Contest:     string(s.Contest),
		ContestName: metaString(s.Meta, "contest_name"),
		Participant: string(s.Participant),
		Problem:     string(s.Problem),
		ProblemName: metaString(s.Meta, "problem_name"),
		Lang:        string(s.Lang),
		SubmittedAt: submitted,
		Verdict:     string(s.Verdict),
		Source:      lines,
	}
}

func formatSignalCount(n int) string {
	mod10 := n % 10
	mod100 := n % 100
	var word string
	switch {
	case mod10 == 1 && mod100 != 11:
		word = "сигнал"
	case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
		word = "сигнала"
	default:
		word = "сигналов"
	}
	return fmt.Sprintf("%d %s", n, word)
}
