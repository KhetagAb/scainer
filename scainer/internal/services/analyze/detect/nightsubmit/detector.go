package nightsubmit

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/detect"
)

const (
	evidenceKind = "night_submit"
)

type Detector struct{}

func NewDetector() *Detector { return &Detector{} }

var _ detect.Detector[domain.StandaloneUnit] = (*Detector)(nil)

func (d *Detector) Name() string { return DetectorName }

func (d *Detector) AI() bool { return false }

func (d *Detector) Analyze(_ context.Context, u domain.StandaloneUnit) ([]domain.Signal, error) {
	sub := u.Sub
	if sub.SubmittedAt.IsZero() || !InNightWindow(sub.SubmittedAt) {
		return nil, nil
	}
	return []domain.Signal{{
		Detector: DetectorName,
		Subject:  domain.NewSubmissionSubject(sub.Contest, sub.ID),
		Score:    1.0,
		Evidence: []domain.Evidence{{
			Kind:        evidenceKind,
			Description: EvidenceDescription(sub.SubmittedAt),
			Spans:       []domain.Span{{Submission: sub.ID}},
		}},
	}}, nil
}
