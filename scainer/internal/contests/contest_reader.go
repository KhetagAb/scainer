package contests

import (
	"context"
	"fmt"
	"slices"

	"scainer/internal/domain"
)

type ContestReader struct {
	registry      ContestRegistry
	store         SubmissionStore
	findingsStore FindingsStore
}

func NewContestReader(registry ContestRegistry, store SubmissionStore, findingsStore FindingsStore) *ContestReader {
	return &ContestReader{
		registry:      registry,
		store:         store,
		findingsStore: findingsStore,
	}
}

func (r *ContestReader) List(ctx context.Context) ([]Contest, error) {
	records, err := r.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(records, func(a, b ContestRecord) int {
		return compareContestID(a.Contest.ID, b.Contest.ID)
	})

	out := make([]Contest, 0, len(records))
	for _, record := range records {
		contest := record.Contest
		if err := r.fillStatistic(ctx, &contest); err != nil {
			return nil, err
		}
		out = append(out, contest)
	}
	return out, nil
}

func (r *ContestReader) fillStatistic(ctx context.Context, contest *Contest) error {
	byProblem, err := r.store.ByProblem(ctx, contest.ID)
	if err != nil {
		return fmt.Errorf("stats ByProblem %s: %w", contest.ID, err)
	}
	submissionCount := 0
	for _, list := range byProblem {
		submissionCount += len(list)
	}
	contest.Statistic.SubmissionCount = submissionCount
	contest.Statistic.ProblemCount = len(byProblem)
	return nil
}

func (r *ContestReader) Problems(ctx context.Context, id domain.ContestID) ([]ProblemInfo, error) {
	record, err := get(ctx, r.registry, id)
	if err != nil {
		return nil, err
	}

	problems, err := r.store.ByProblem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get problems: %w", err)
	}

	excluded := make(map[domain.ProblemID]bool)
	for _, problemID := range record.Contest.ExcludedProblems {
		excluded[problemID] = true
	}

	out := make([]ProblemInfo, 0, len(problems))
	for problemID, submissions := range problems {
		problemName := ""
		if len(submissions) > 0 && len(submissions[0].Meta) > 0 {
			if name, ok := submissions[0].Meta["problem_name"]; ok {
				if nameStr, ok := name.(string); ok {
					problemName = nameStr
				}
			}
		}
		pending := 0
		for _, sub := range submissions {
			if sub.Verdict == domain.VerdictPR {
				pending++
			}
		}

		out = append(out, ProblemInfo{
			ID:              problemID,
			Name:            problemName,
			Excluded:        excluded[problemID],
			SubmissionCount: len(submissions),
			PendingCount:    pending,
		})
	}
	return out, nil
}

func (r *ContestReader) GetFindings(ctx context.Context, id domain.ContestID) ([]domain.Finding, map[domain.SubmissionID]domain.Submission, error) {
	if _, err := get(ctx, r.registry, id); err != nil {
		return nil, nil, err
	}

	snapshot, ok, err := r.findingsStore.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, nil
	}

	submissions, err := loadReferencedSubmissions(ctx, r.store, snapshot.Findings)
	if err != nil {
		return nil, nil, err
	}
	return snapshot.Findings, submissions, nil
}
