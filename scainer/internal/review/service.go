package review

import (
	"context"
	"errors"
	"sort"
	"strings"

	"scainer/internal/contests"
	"scainer/internal/domain"
)

var ErrInvalidVerdict = errors.New("review: verdict must be OK or RJ")

type Service struct {
	store    contests.SubmissionStore
	comments CommentsProvider
	status   StatusSyncer
}

func New(store contests.SubmissionStore, comments CommentsProvider, status StatusSyncer) *Service {
	return &Service{store: store, comments: comments, status: status}
}

func (s *Service) List(ctx context.Context, contest domain.ContestID) ([]SubmissionReviewItem, error) {
	byProblem, err := s.store.ByProblem(ctx, contest)
	if err != nil {
		return nil, err
	}
	out := make([]SubmissionReviewItem, 0)
	for _, subs := range byProblem {
		for _, sub := range subs {
			out = append(out, SubmissionReviewItem{
				ID:          sub.ID,
				Problem:     sub.Problem,
				Participant: sub.Participant,
				Lang:        sub.Lang,
				SubmittedAt: sub.SubmittedAt,
				Verdict:     sub.Verdict,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].SubmittedAt.Before(out[j].SubmittedAt)
	})
	return out, nil
}

func (s *Service) LoadComments(ctx context.Context, contest domain.ContestID, submissionID domain.SubmissionID) (CommentsResult, error) {
	sub, err := s.store.GetByID(ctx, submissionID)
	if err != nil {
		return CommentsResult{}, err
	}

	res := CommentsResult{
		Verdict: sub.Verdict,
		Source:  strings.Split(string(sub.Source), "\n"),
	}
	updated, syncErr := s.status.Sync(ctx, sub)
	if syncErr != nil {
		res.StatusStale = true
		res.StatusError = syncErr.Error()
	} else {
		res.Verdict = updated.Verdict
		_ = s.store.Put(ctx, []domain.Submission{updated})
		sub = updated
	}

	msgs, listErr := s.comments.List(ctx, sub)
	if listErr != nil {
		res.CommentsError = listErr.Error()
		res.Comments = nil
		return res, nil
	}
	res.Comments = msgs
	return res, nil
}

func (s *Service) Comment(ctx context.Context, contest domain.ContestID, submissionID domain.SubmissionID, text string) error {
	sub, err := s.store.GetByID(ctx, submissionID)
	if err != nil {
		return err
	}
	return s.comments.Post(ctx, sub, text)
}

func (s *Service) Decide(ctx context.Context, contest domain.ContestID, submissionID domain.SubmissionID, req DecideRequest) error {
	if req.Verdict != domain.VerdictOK && req.Verdict != domain.VerdictRJ {
		return ErrInvalidVerdict
	}
	sub, err := s.store.GetByID(ctx, submissionID)
	if err != nil {
		return err
	}

	if req.Comment != "" {
		if err := s.comments.Post(ctx, sub, req.Comment); err != nil {
			return err
		}
	}

	updated, err := s.status.SetVerdict(ctx, sub, req.Verdict)
	if err != nil {
		return err
	}
	return s.store.Put(ctx, []domain.Submission{updated})
}
