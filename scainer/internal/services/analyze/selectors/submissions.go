package selectors

import (
	"context"
	"sort"

	"scainer/internal/domain"
)

func Submissions(contest domain.ContestID) Selector[domain.StandaloneUnit] {
	return func(ctx context.Context, st Store) ([]domain.StandaloneUnit, error) {
		byProblem, err := st.ByProblem(ctx, contest)
		if err != nil {
			return nil, err
		}

		problems := make([]domain.ProblemID, 0, len(byProblem))
		for p := range byProblem {
			problems = append(problems, p)
		}
		sort.Slice(problems, func(i, j int) bool { return problems[i] < problems[j] })

		var units []domain.StandaloneUnit
		for _, p := range problems {
			for _, sub := range byProblem[p] {
				units = append(units, domain.StandaloneUnit{Sub: sub})
			}
		}
		return units, nil
	}
}
