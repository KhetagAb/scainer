package detect

import (
	"context"

	"scainer/internal/domain"
)

type Detector[U domain.Unit] interface {
	Name() string
	AI() bool
	Analyze(ctx context.Context, u U) ([]domain.Signal, error)
}

// Process-wide лимит одновременных Detector.Analyze: один инстанс на все job'ы,
// иначе параллельные job'ы перемножили бы число живых JVM (JPlag).
type Limiter struct {
	sem chan struct{}
}

func NewLimiter(n int) *Limiter {
	if n < 1 {
		n = 1
	}
	return &Limiter{sem: make(chan struct{}, n)}
}

func (l *Limiter) Acquire(ctx context.Context) error {
	return l.acquire(ctx)
}

func (l *Limiter) Release() { l.release() }

func (l *Limiter) acquire(ctx context.Context) error {
	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limiter) release() { <-l.sem }
