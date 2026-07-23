package ejudge

import (
	"context"
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/review"
)

type Comments struct {
	ClientResolver ClientResolver
}

var _ review.CommentsProvider = (*Comments)(nil)

func NewComments(clientResolver ClientResolver) *Comments {
	return &Comments{ClientResolver: clientResolver}
}

func (p *Comments) List(ctx context.Context, sub domain.Submission) ([]review.Comment, error) {
	client, err := clientFor(ctx, p.ClientResolver)
	if err != nil {
		return nil, err
	}
	key, err := parseSubmissionKey(sub.ID)
	if err != nil {
		return nil, err
	}
	msgs, err := client.RunMessages(ctx, key.ContestID, key.RunID)
	if err != nil {
		return nil, err
	}
	out := make([]review.Comment, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, review.Comment{
			ID:      fmt.Sprintf("%d", m.ClarID),
			From:    m.From,
			Subject: m.Subject,
			Text:    m.Text,
			Time:    m.Time,
		})
	}
	return out, nil
}

func (p *Comments) Post(ctx context.Context, sub domain.Submission, text string) error {
	client, err := clientFor(ctx, p.ClientResolver)
	if err != nil {
		return err
	}
	key, err := parseSubmissionKey(sub.ID)
	if err != nil {
		return err
	}
	return client.SendRunComment(ctx, key.ContestID, key.RunID, text)
}
