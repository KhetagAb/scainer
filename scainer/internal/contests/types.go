package contests

import (
	"cmp"
	"context"
	"errors"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
)

var ErrContestNotFound = errors.New("contest not found")
var ErrDuplicateContest = errors.New("contest already registered")

type SourceSpec struct {
	Type   string
	Config yaml.Node
}

type Registration struct {
	ID               domain.ContestID
	ParallelID       string
	Source           *SourceSpec
	ExcludedProblems []domain.ProblemID
}

type Contest struct {
	ID               domain.ContestID   `bson:"id"`
	Name             string             `bson:"name"`
	ParallelID       string             `bson:"parallel_id"`
	ExcludedProblems []domain.ProblemID `bson:"excluded_problems"`
	LastImportedAt   *time.Time         `bson:"last_imported_at,omitempty"`
	Statistic        ContestStatistic   `bson:"-"`
}

type ContestStatistic struct {
	SubmissionCount int
	ProblemCount    int
}

type ProblemInfo struct {
	ID              domain.ProblemID
	Name            string
	Excluded        bool
	SubmissionCount int
	PendingCount    int
}

type FindingsSnapshot struct {
	ContestID  domain.ContestID `bson:"contest_id"`
	Findings   []domain.Finding `bson:"findings"`
	ComputedAt time.Time        `bson:"computed_at"`
}

type SubmissionStore interface {
	Put(ctx context.Context, submissions []domain.Submission) error
	GetByID(ctx context.Context, id domain.SubmissionID) (domain.Submission, error)
	GetByIDs(ctx context.Context, ids []domain.SubmissionID) ([]domain.Submission, error)
	ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error)
	GetCursor(ctx context.Context, key string) (value string, ok bool, err error)
	SetCursor(ctx context.Context, key string, value string) error
}

type FindingsStore interface {
	Put(ctx context.Context, snapshot FindingsSnapshot) error
	Get(ctx context.Context, id domain.ContestID) (FindingsSnapshot, bool, error)
	Delete(ctx context.Context, id domain.ContestID) error
}

func get(ctx context.Context, registry ContestRegistry, id domain.ContestID) (ContestRecord, error) {
	record, ok, err := registry.Get(ctx, id)
	if err != nil {
		return ContestRecord{}, err
	}
	if !ok {
		return ContestRecord{}, ErrContestNotFound
	}
	return record, nil
}

func compareContestID(a, b domain.ContestID) int {
	ai, aErr := strconv.Atoi(string(a))
	bi, bErr := strconv.Atoi(string(b))
	if aErr == nil && bErr == nil {
		return cmp.Compare(ai, bi)
	}
	return cmp.Compare(string(a), string(b))
}

func loadReferencedSubmissions(ctx context.Context, store SubmissionStore, findings []domain.Finding) (map[domain.SubmissionID]domain.Submission, error) {
	seen := make(map[domain.SubmissionID]bool)
	var ids []domain.SubmissionID
	for _, finding := range findings {
		for _, signal := range finding.Signals {
			for _, evidence := range signal.Evidence {
				for _, span := range evidence.Spans {
					if seen[span.Submission] {
						continue
					}
					seen[span.Submission] = true
					ids = append(ids, span.Submission)
				}
			}
		}
	}
	list, err := store.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[domain.SubmissionID]domain.Submission, len(list))
	for _, submission := range list {
		out[submission.ID] = submission
	}
	return out, nil
}
