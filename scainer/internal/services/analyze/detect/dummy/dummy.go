package dummy

import (
	"context"

	"scainer/internal/services/analyze/detect"
	"scainer/internal/domain"
)

type AlwaysProblem struct{}

func (AlwaysProblem) Name() string { return "always_problem" }

func (AlwaysProblem) AI() bool { return false }

func (AlwaysProblem) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	return pairSignals(u, "always_problem", 1.0, "заглушка: всегда флажит пару"), nil
}

type AlwaysClean struct{}

func (AlwaysClean) Name() string { return "always_clean" }

func (AlwaysClean) AI() bool { return false }

func (AlwaysClean) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	return pairSignals(u, "always_clean", 0.0, "заглушка: всё чисто"), nil
}

var (
	_ detect.Detector[domain.ProblemUnit] = AlwaysProblem{}
	_ detect.Detector[domain.ProblemUnit] = AlwaysClean{}
)

func pairSignals(u domain.ProblemUnit, detector string, score float64, why string) []domain.Signal {
	var out []domain.Signal
	for i := 0; i < len(u.Subs); i++ {
		for j := i + 1; j < len(u.Subs); j++ {
			out = append(out, domain.Signal{
				Detector: detector,
				Subject: domain.NewPairSubject(
					u.Subs[i].Contest,
					u.Problem,
					u.Subs[i].Participant,
					u.Subs[j].Participant,
				),
				Score:    score,
				Evidence: []domain.Evidence{{Kind: "dummy", Description: why}},
			})
		}
	}
	return out
}
