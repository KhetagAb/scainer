package detect

import (
	"context"
	"maps"
	"slices"

	"scainer/internal/domain"
)

const OkWithLastMin = 4

type OkWithLastSelector struct {
	Contest domain.ContestID
}

var _ Selector[domain.ProblemParticipantUnit] = OkWithLastSelector{}

func (s OkWithLastSelector) Select(ctx context.Context, store Store) ([]domain.ProblemParticipantUnit, error) {
	byProblem, err := store.ByProblem(ctx, s.Contest)
	if err != nil {
		return nil, err
	}

	problems := slices.Sorted(maps.Keys(byProblem))
	var units []domain.ProblemParticipantUnit
	for _, problem := range problems {
		byParticipant := make(map[domain.ParticipantID][]domain.Submission)
		for _, sub := range byProblem[problem] {
			byParticipant[sub.Participant] = append(byParticipant[sub.Participant], sub)
		}
		for _, participant := range slices.Sorted(maps.Keys(byParticipant)) {
			window := WindowEndingAtLastOK(byParticipant[participant], OkWithLastMin)
			if len(window) == 0 {
				continue
			}
			units = append(units, domain.ProblemParticipantUnit{
				Problem:     problem,
				Participant: participant,
				Subs:        window,
			})
		}
	}
	return units, nil
}
