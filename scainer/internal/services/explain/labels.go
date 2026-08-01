package explain

import (
	"context"

	"scainer/internal/domain"
)

type LabelResolver interface {
	ProblemLabel(ctx context.Context, contestID domain.ContestID, problemID domain.ProblemID) (string, bool, error)
}

type submissionByProblemStore interface {
	ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error)
}

type submissionLabelResolver struct {
	store submissionByProblemStore
}

func NewSubmissionLabelResolver(store submissionByProblemStore) LabelResolver {
	return &submissionLabelResolver{store: store}
}

func (r *submissionLabelResolver) ProblemLabel(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (string, bool, error) {
	byProblem, err := r.store.ByProblem(ctx, contestID)
	if err != nil {
		return "", false, err
	}
	label, ok := problemLabelFromSubmissions(byProblem, problemID)
	return label, ok, nil
}

func problemLabelFromSubmissions(byProblem map[domain.ProblemID][]domain.Submission, problemID domain.ProblemID) (string, bool) {
	subs := byProblem[problemID]
	if len(subs) == 0 {
		return "", false
	}
	for _, sub := range subs {
		if sub.Meta == nil {
			continue
		}
		if name, ok := sub.Meta["problem_name"].(string); ok && name != "" {
			return name, true
		}
	}
	return "", false
}
