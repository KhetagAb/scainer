package ejudge

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/review"
	ejudgeapi "scainer/pkg/ejudge"
)

type Status struct {
	ClientResolver ClientResolver
}

var _ review.StatusProvider = (*Status)(nil)

func NewStatus(clientResolver ClientResolver) *Status {
	return &Status{ClientResolver: clientResolver}
}

func (s *Status) Sync(ctx context.Context, sub domain.Submission) (domain.Submission, error) {
	client, err := clientFor(ctx, s.ClientResolver)
	if err != nil {
		return domain.Submission{}, err
	}
	key, err := parseSubmissionKey(sub.ID)
	if err != nil {
		return domain.Submission{}, err
	}
	info, err := client.RunStatus(ctx, key.ContestID, key.RunID)
	if err != nil {
		return domain.Submission{}, err
	}
	return submissionWithVerdict(sub, toDomainVerdict(info.Verdict), info.Status), nil
}

func (s *Status) SetVerdict(ctx context.Context, sub domain.Submission, verdict domain.Verdict) (domain.Submission, error) {
	client, err := clientFor(ctx, s.ClientResolver)
	if err != nil {
		return domain.Submission{}, err
	}
	key, err := parseSubmissionKey(sub.ID)
	if err != nil {
		return domain.Submission{}, err
	}
	ev, ok := toEjudgeVerdict(verdict)
	if !ok {
		return domain.Submission{}, fmt.Errorf("ejudge review: неизвестный вердикт %q", verdict)
	}
	if err := client.ChangeRunStatus(ctx, key.ContestID, key.RunID, ev); err != nil {
		return domain.Submission{}, err
	}
	code, ok := ejudgeapi.StatusCode(ev)
	if !ok {
		return domain.Submission{}, fmt.Errorf("ejudge review: неизвестный вердикт %q", verdict)
	}
	return submissionWithVerdict(sub, verdict, code), nil
}

func submissionWithVerdict(sub domain.Submission, verdict domain.Verdict, statusCode int) domain.Submission {
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
