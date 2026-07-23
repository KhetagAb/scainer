package review

import (
	"context"
	"time"

	"scainer/internal/domain"
)

type SubmissionReviewItem struct {
	ID          domain.SubmissionID
	Problem     domain.ProblemID
	Participant domain.ParticipantID
	Lang        domain.Lang
	SubmittedAt time.Time
	Verdict     domain.Verdict
}

type Comment struct {
	ID      string
	From    string // display name
	Subject string
	Text    string
	Time    time.Time
}

type CommentsResult struct {
	Comments      []Comment
	Source        []string
	Verdict       domain.Verdict
	StatusStale   bool
	StatusError   string
	CommentsError string
}

type DecideRequest struct {
	Verdict domain.Verdict
	Comment string
}

type StatusProvider interface {
	Sync(ctx context.Context, sub domain.Submission) (domain.Submission, error)
	SetVerdict(ctx context.Context, sub domain.Submission, verdict domain.Verdict) (domain.Submission, error)
}

type CommentsProvider interface {
	List(ctx context.Context, sub domain.Submission) ([]Comment, error)
	Post(ctx context.Context, sub domain.Submission, text string) error
}
