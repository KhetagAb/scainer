package detect

import (
	"context"
	"sort"

	"scainer/internal/domain"
)

type ProblemSelector struct {
	Contest domain.ContestID
}

var _ Selector[domain.ProblemUnit] = ProblemSelector{}

func (s ProblemSelector) Select(ctx context.Context, st Store) ([]domain.ProblemUnit, error) {
	byProblem, err := st.ByProblem(ctx, s.Contest)
	if err != nil {
		return nil, err
	}

	problems := make([]domain.ProblemID, 0, len(byProblem))
	for p := range byProblem {
		problems = append(problems, p)
	}
	sort.Slice(problems, func(i, j int) bool { return problems[i] < problems[j] })

	var units []domain.ProblemUnit
	for _, p := range problems {
		byLang := make(map[domain.Lang][]domain.Submission)
		var langs []domain.Lang
		for _, sub := range byProblem[p] {
			if _, ok := byLang[sub.Lang]; !ok {
				langs = append(langs, sub.Lang)
			}
			byLang[sub.Lang] = append(byLang[sub.Lang], sub)
		}
		sort.Slice(langs, func(i, j int) bool { return langs[i] < langs[j] })
		for _, lang := range langs {
			units = append(units, domain.ProblemUnit{Problem: p, Lang: lang, Subs: byLang[lang]})
		}
	}
	return units, nil
}

type SubmissionsSelector struct {
	Contest domain.ContestID
}

var _ Selector[domain.StandaloneUnit] = SubmissionsSelector{}

func (s SubmissionsSelector) Select(ctx context.Context, st Store) ([]domain.StandaloneUnit, error) {
	byProblem, err := st.ByProblem(ctx, s.Contest)
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
