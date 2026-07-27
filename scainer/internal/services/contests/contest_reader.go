package contests

import (
	"context"
	"fmt"
	"slices"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/scoring"
)

type ContestReader struct {
	registry     ContestRegistry
	store        SubmissionStore
	analysisRepo AnalysisRepository
	scorer       scoring.Scorer
}

func NewContestReader(
	registry ContestRegistry,
	store SubmissionStore,
	analysisRepo AnalysisRepository,
	scorer scoring.Scorer,
) *ContestReader {
	return &ContestReader{
		registry:     registry,
		store:        store,
		analysisRepo: analysisRepo,
		scorer:       scorer,
	}
}

func (r *ContestReader) List(ctx context.Context) ([]ContestSummary, error) {
	records, err := r.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(records, func(a, b ContestRecord) int {
		return compareContestID(a.Contest.ID, b.Contest.ID)
	})

	out := make([]ContestSummary, 0, len(records))
	for _, record := range records {
		summary, err := r.buildSummary(ctx, record.Contest)
		if err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	return out, nil
}

func (r *ContestReader) Summary(ctx context.Context, id domain.ContestID) (ContestSummary, error) {
	record, err := get(ctx, r.registry, id)
	if err != nil {
		return ContestSummary{}, err
	}
	return r.buildSummary(ctx, record.Contest)
}

func (r *ContestReader) buildSummary(ctx context.Context, contest Contest) (ContestSummary, error) {
	submissionCount, problemCount, err := r.countSubmissions(ctx, contest.ID)
	if err != nil {
		return ContestSummary{}, err
	}

	computedAt, err := r.computedAt(ctx, contest.ID)
	if err != nil {
		return ContestSummary{}, err
	}

	return ContestSummary{
		ID:              contest.ID,
		Name:            contest.Name,
		ParallelID:      contest.ParallelID,
		LastImportedAt:  contest.LastImportedAt,
		ComputedAt:      computedAt,
		SubmissionCount: submissionCount,
		ProblemCount:    problemCount,
	}, nil
}

func (r *ContestReader) computedAt(ctx context.Context, id domain.ContestID) (*time.Time, error) {
	snap, ok, err := r.analysisRepo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("analysis snapshot %s: %w", id, err)
	}
	if !ok || snap.ComputedAt.IsZero() {
		return nil, nil
	}
	t := snap.ComputedAt
	return &t, nil
}

func (r *ContestReader) countSubmissions(ctx context.Context, id domain.ContestID) (submissionCount, problemCount int, err error) {
	byProblem, err := r.store.ByProblem(ctx, id)
	if err != nil {
		return 0, 0, fmt.Errorf("stats ByProblem %s: %w", id, err)
	}
	for _, list := range byProblem {
		submissionCount += len(list)
	}
	return submissionCount, len(byProblem), nil
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

	snapshot, ok, err := r.analysisRepo.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, nil
	}

	findings := snapshot.Findings(r.scorer)
	submissions, err := loadReferencedSubmissions(ctx, r.store, findings)
	if err != nil {
		return nil, nil, err
	}
	return findings, submissions, nil
}
