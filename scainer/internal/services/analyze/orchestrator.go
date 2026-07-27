package analyze

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/analyze/selectors"
	"scainer/internal/services/contests"
	"scainer/pkg/progress"
)

func NewOrchestrator(registry *Registry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

func (o *Orchestrator) Run(
	ctx context.Context,
	contestID domain.ContestID,
	store selectors.Store,
	prev contests.AnalysisSnapshot,
) (contests.AnalysisSnapshot, error) {
	next := prev.Clone()
	next.ContestID = contestID

	in := contestRun{ctx: ctx, contestID: contestID, store: store}
	var aggErr error
	for _, run := range o.registry.Runs() {
		if err := run(in, &next); err != nil {
			aggErr = errors.Join(aggErr, err)
		}
	}

	next.ComputedAt = time.Now().UTC()
	return next, aggErr
}

type detectorRun[U domain.Unit] struct {
	name     string
	det      detect.Detector[U]
	policy   InvalidationPolicy[U]
	selector func(domain.ContestID) selectors.Selector[U]
	limiter  *detect.Limiter
}

func (d *detectorRun[U]) run(in contestRun, snap *contests.AnalysisSnapshot) error {
	units, err := d.selector(in.contestID)(in.ctx, in.store)
	if err != nil {
		return fmt.Errorf("%s: %w", d.name, err)
	}

	prog, sigs := snap.DetectorMaps(d.name)
	todo := scheduleUnits(in.ctx, units, d.policy, prog)
	if len(todo) == 0 {
		return nil
	}

	runner := &unitRunner[U]{
		detectorRun: d,
		in:          in,
		progress:    &prog,
		signals:     &sigs,
		total:       len(todo),
	}
	if err := runner.all(todo); err != nil {
		return fmt.Errorf("%s: %w", d.name, err)
	}
	if runner.err != nil {
		return fmt.Errorf("%s: %w", d.name, runner.err)
	}
	return nil
}

type scheduledUnit[U domain.Unit] struct {
	unit U
	key  contests.ScopeKey
}

func scheduleUnits[U domain.Unit](
	ctx context.Context,
	units []U,
	policy InvalidationPolicy[U],
	prev contests.DetectorProgress,
) []scheduledUnit[U] {
	var todo []scheduledUnit[U]
	for _, u := range units {
		if !policy.ShouldRun(ctx, u, prev) {
			continue
		}
		todo = append(todo, scheduledUnit[U]{unit: u, key: policy.ScopeKey(u)})
	}
	return todo
}

type unitRunner[U domain.Unit] struct {
	*detectorRun[U]
	in       contestRun
	progress *contests.DetectorProgress
	signals  *contests.DetectorSignals
	total    int

	mu   sync.Mutex
	done int
	err  error
}

func (r *unitRunner[U]) all(todo []scheduledUnit[U]) error {
	progress.Report(r.in.ctx, progress.Event{Phase: "analyzing", Done: 0, Total: r.total})

	g, ctx := errgroup.WithContext(r.in.ctx)
	for _, job := range todo {
		g.Go(func() error {
			return r.runOne(ctx, job)
		})
	}
	return g.Wait()
}

func (r *unitRunner[U]) runOne(ctx context.Context, job scheduledUnit[U]) error {
	if err := r.limiter.Acquire(ctx); err != nil {
		return err
	}
	defer r.limiter.Release()

	sigs, err := r.det.Analyze(ctx, job.unit)
	if err != nil {
		r.onError(job.key, err)
		return nil
	}
	r.onSuccess(job, sigs)
	return nil
}

func (r *unitRunner[U]) onError(key contests.ScopeKey, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.err = errors.Join(r.err, fmt.Errorf("%s[%s]: %w", r.det.Name(), key, err))
	r.tick()
}

func (r *unitRunner[U]) onSuccess(job scheduledUnit[U], sigs []domain.Signal) {
	if ai := r.det.AI(); ai {
		for i := range sigs {
			sigs[i].AI = true
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	(*r.signals)[job.key] = sigs
	r.policy.AfterRun(job.unit, r.progress)
	r.tick()
}

func (r *unitRunner[U]) tick() {
	r.done++
	progress.Report(r.in.ctx, progress.Event{Phase: "analyzing", Done: r.done, Total: r.total})
}
