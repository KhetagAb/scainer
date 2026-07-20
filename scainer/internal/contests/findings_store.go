package contests

import (
	"context"
	"time"

	"github.com/lksh/scainer/internal/domain"
)

type FindingsSnapshot struct {
	ContestID  domain.ContestID `bson:"contest_id"`
	Findings   []domain.Finding `bson:"findings"`
	ComputedAt time.Time        `bson:"computed_at"`
}

type FindingsStore interface {
	Put(ctx context.Context, snap FindingsSnapshot) error
	Get(ctx context.Context, id domain.ContestID) (FindingsSnapshot, bool, error)
	Delete(ctx context.Context, id domain.ContestID) error
}
