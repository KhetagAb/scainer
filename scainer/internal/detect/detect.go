package detect

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"scainer/internal/domain"
	"scainer/internal/progress"
	"scainer/internal/store"
)

type Detector[U domain.Unit] interface {
	Name() string
	AI() bool
	Analyze(ctx context.Context, u U) ([]domain.Signal, error)
}

type Selector[U domain.Unit] interface {
	Select(ctx context.Context, s store.Store, policy AnalysisPolicy) ([]U, error)
}

type Stage interface {
	Run(ctx context.Context, s store.Store, policy AnalysisPolicy) ([]domain.Signal, error)
}

// Process-wide лимит одновременных Detector.Analyze: один инстанс на все Stage/job'ы,
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

func (l *Limiter) acquire(ctx context.Context) error {
	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limiter) release() { <-l.sem }

func NewStage[U domain.Unit](sel Selector[U], limiter *Limiter, dets ...Detector[U]) Stage {
	return stage[U]{sel: sel, limiter: limiter, dets: dets}
}

type stage[U domain.Unit] struct {
	sel     Selector[U]
	limiter *Limiter
	dets    []Detector[U]
}

func (st stage[U]) Run(ctx context.Context, s store.Store, policy AnalysisPolicy) ([]domain.Signal, error) {
	units, err := st.sel.Select(ctx, s, policy)
	if err != nil {
		return nil, err
	}

	total := len(units) * len(st.dets)
	var (
		mu   sync.Mutex
		out  []domain.Signal
		done int
	)

	g, gctx := errgroup.WithContext(ctx)
	for _, u := range units {
		for _, d := range st.dets {
			u, d := u, d
			g.Go(func() error {
				if err := st.limiter.acquire(gctx); err != nil {
					return err
				}
				defer st.limiter.release()

				// TODO(step): фильтрация детекторов по policy.DetectorsByProblem с учётом задачи юнита.
				sigs, err := d.Analyze(gctx, u)
				if err != nil {
					return err
				}
				ai := d.AI()
				for j := range sigs {
					sigs[j].AI = ai
				}

				mu.Lock()
				out = append(out, sigs...)
				done++
				progress.Report(ctx, progress.Event{Phase: "analyzing", Done: done, Total: total})
				mu.Unlock()
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return out, nil
}
