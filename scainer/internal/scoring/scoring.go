package scoring

import (
	"sort"

	"github.com/lksh/scainer/internal/domain"
)

type Scorer interface {
	Score(signals []domain.Signal) []domain.Finding
}

// TODO(step): настраиваемые веса детекторов; сборка кластеров (>2 похожих) из попарных сигналов.
type Weighted struct{}

func NewWeighted() Weighted { return Weighted{} }

var _ Scorer = Weighted{}

func (Weighted) Score(signals []domain.Signal) []domain.Finding {
	type agg struct {
		subject domain.Subject
		sum     float64
		count   int
		signals []domain.Signal
	}
	groups := make(map[string]*agg)
	var order []string

	for _, sig := range signals {
		key := domain.SubjectKey(sig.Subject)
		g, ok := groups[key]
		if !ok {
			g = &agg{subject: sig.Subject}
			groups[key] = g
			order = append(order, key)
		}
		g.sum += sig.Score
		g.count++
		g.signals = append(g.signals, sig)
	}

	findings := make([]domain.Finding, 0, len(groups))
	for _, key := range order {
		g := groups[key]
		score := 0.0
		if g.count > 0 {
			score = g.sum / float64(g.count)
		}
		findings = append(findings, domain.Finding{Subject: g.subject, Score: score, Signals: g.signals})
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Score != findings[j].Score {
			return findings[i].Score > findings[j].Score
		}
		return domain.SubjectKey(findings[i].Subject) < domain.SubjectKey(findings[j].Subject)
	})
	return findings
}
