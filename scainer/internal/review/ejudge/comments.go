package ejudge

import (
	"context"
	"fmt"
	"strconv"

	"scainer/internal/domain"
	"scainer/internal/review"
	ejudgeapi "scainer/pkg/ejudge"
)

type Comments struct {
	Client *ejudgeapi.Client
}

var _ review.CommentsProvider = (*Comments)(nil)

func (p *Comments) List(ctx context.Context, sub domain.Submission) ([]review.Comment, error) {
	if p == nil || p.Client == nil {
		return nil, fmt.Errorf("review/ejudge: клиент не задан")
	}
	key, err := ejudgeapi.ParseSubmissionID(sub.ID)
	if err != nil {
		return nil, err
	}
	msgs, err := p.Client.RunMessages(ctx, key.ContestID, key.RunID)
	if err != nil {
		return nil, err
	}
	out := make([]review.Comment, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, review.Comment{
			ID:      strconv.Itoa(m.ClarID),
			From:    m.From,
			Subject: m.Subject,
			Text:    m.Text,
			Time:    m.Time,
		})
	}
	return out, nil
}

func (p *Comments) Post(ctx context.Context, sub domain.Submission, text string) error {
	if p == nil || p.Client == nil {
		return fmt.Errorf("review/ejudge: клиент не задан")
	}
	key, err := ejudgeapi.ParseSubmissionID(sub.ID)
	if err != nil {
		return err
	}
	return p.Client.SendRunComment(ctx, key.ContestID, key.RunID, text)
}
