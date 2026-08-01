package ejudge

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/review"
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
	return submissionWithVerdict(sub, mapToDomainVerdict(info.Status, info.StatusStr), info.Status, info.StatusStr), nil
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
	code, ok := domainVerdictToStatusCode(verdict)
	if !ok {
		return domain.Submission{}, fmt.Errorf("ejudge review: неизвестный вердикт %q", verdict)
	}
	if err := client.ChangeRunStatus(ctx, key.ContestID, key.RunID, code); err != nil {
		return domain.Submission{}, err
	}
	return submissionWithVerdict(sub, verdict, code, string(verdict)), nil
}

func submissionWithVerdict(sub domain.Submission, verdict domain.Verdict, statusCode int, ejudgeStatusStr string) domain.Submission {
	out := sub
	out.Verdict = verdict
	if out.Meta == nil {
		out.Meta = make(map[string]any, 3)
	} else {
		meta := make(map[string]any, len(sub.Meta)+3)
		for k, v := range sub.Meta {
			meta[k] = v
		}
		out.Meta = meta
	}
	out.Meta["status"] = statusCode
	if ejudgeStatusStr != "" {
		out.Meta["ejudge_status_str"] = ejudgeStatusStr
		out.Meta["status_str"] = ejudgeStatusStr
	} else {
		out.Meta["status_str"] = string(verdict)
	}
	return out
}
