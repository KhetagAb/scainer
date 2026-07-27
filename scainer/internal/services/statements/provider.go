package statements

import (
	"context"
	"io"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

type Contest struct {
	ID         domain.ContestID
	ParallelID string
	Source     contests.SourceSpec
}

type Document struct {
	Body        io.ReadCloser
	ContentType string
	Filename    string
}

type Provider interface {
	SourceType() string
	Fetch(ctx context.Context, c Contest) (Document, error)
}
