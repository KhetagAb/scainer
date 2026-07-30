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
	byProblem, err := r.store.ByProblem(ctx, contest.ID)
	if err != nil {
		return ContestSummary{}, fmt.Errorf("stats ByProblem %s: %w", contest.ID, err)
	}

	submissionCount := 0
	for _, list := range byProblem {
		submissionCount += len(list)
	}

	computedAt, err := r.computedAt(ctx, contest.ID)
	if err != nil {
		return ContestSummary{}, err
	}

	problems := buildProblemInfos(contest, byProblem)

	return ContestSummary{
		ID:              contest.ID,
		Name:            contest.Name,
		ParallelID:      contest.ParallelID,
		LastImportedAt:  contest.LastImportedAt,
		ComputedAt:      computedAt,
		SubmissionCount: submissionCount,
		Problems:        problems,
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

func (r *ContestReader) Problems(ctx context.Context, id domain.ContestID) ([]ProblemInfo, error) {
	record, err := get(ctx, r.registry, id)
	if err != nil {
		return nil, err
	}

	byProblem, err := r.store.ByProblem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get problems: %w", err)
	}

	return buildProblemInfos(record.Contest, byProblem), nil
}

func buildProblemInfos(contest Contest, byProblem map[domain.ProblemID][]domain.Submission) []ProblemInfo {
	out := make([]ProblemInfo, 0, len(contest.Problems))
	for _, problem := range contest.Problems {
		submissions := byProblem[problem.ID]
		pending := 0
		for _, sub := range submissions {
			if sub.Verdict == domain.VerdictPR {
				pending++
			}
		}
		out = append(out, ProblemInfo{
			ID:              problem.ID,
			Name:            problem.Name,
			SubmissionCount: len(submissions),
			PendingCount:    pending,
		})
	}
	return out
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
