package selectors

import (
	"context"
	"maps"
	"slices"

	"scainer/internal/domain"
)

// OkWithLastWindowSize — сколько последних посылок до last OK берёт OkWithLast на участника.
const OkWithLastWindowSize = 4

func OkWithLast(contest domain.ContestID) Selector[domain.ProblemParticipantUnit] {
	return func(ctx context.Context, store Store) ([]domain.ProblemParticipantUnit, error) {
		byProblem, err := store.ByProblem(ctx, contest)
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
				window := WindowEndingAtLastOK(byParticipant[participant], OkWithLastWindowSize)
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
}
