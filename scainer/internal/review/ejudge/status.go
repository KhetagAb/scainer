package ejudge

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/review"
	ejudgeapi "scainer/pkg/ejudge"
)

type Status struct {
	Client *ejudgeapi.Client
}

var _ review.StatusSyncer = (*Status)(nil)

func (s *Status) Sync(ctx context.Context, sub domain.Submission) (domain.Submission, error) {
	if s == nil || s.Client == nil {
		return domain.Submission{}, fmt.Errorf("review/ejudge: клиент не задан")
	}
	key, err := ejudgeapi.ParseSubmissionID(sub.ID)
	if err != nil {
		return domain.Submission{}, err
	}
	info, err := s.Client.RunStatus(ctx, key.ContestID, key.RunID)
	if err != nil {
		return domain.Submission{}, err
	}
	return withVerdict(sub, info.Verdict, info.Status), nil
}

func (s *Status) SetVerdict(ctx context.Context, sub domain.Submission, verdict domain.Verdict) (domain.Submission, error) {
	if s == nil || s.Client == nil {
		return domain.Submission{}, fmt.Errorf("review/ejudge: клиент не задан")
	}
	key, err := ejudgeapi.ParseSubmissionID(sub.ID)
	if err != nil {
		return domain.Submission{}, err
	}
	if err := s.Client.ChangeRunStatus(ctx, key.ContestID, key.RunID, verdict); err != nil {
		return domain.Submission{}, err
	}
	code, ok := ejudgeapi.StatusCode(verdict)
	if !ok {
		return domain.Submission{}, fmt.Errorf("review/ejudge: неизвестный вердикт %q", verdict)
	}
	return withVerdict(sub, verdict, code), nil
}

func withVerdict(sub domain.Submission, verdict domain.Verdict, statusCode int) domain.Submission {
	out := sub
	out.Verdict = verdict
	if out.Meta == nil {
		out.Meta = make(map[string]any, 2)
	} else {
		meta := make(map[string]any, len(sub.Meta)+2)
		for k, v := range sub.Meta {
			meta[k] = v
		}
		out.Meta = meta
	}
	out.Meta["status"] = statusCode
	out.Meta["status_str"] = string(verdict)
	return out
}
