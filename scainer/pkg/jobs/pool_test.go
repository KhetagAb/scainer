package jobs_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"scainer/pkg/jobs"
	"scainer/internal/progress"
)

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}

func waitTerminal(t *testing.T, p *jobs.Pool, id string) jobs.State {
	t.Helper()
	var st jobs.State
	waitUntil(t, 2*time.Second, func() bool {
		s, ok := p.Get(id)
		if !ok {
			t.Fatalf("job %s disappeared", id)
		}
		st = s
		return s.Status == jobs.StatusSucceeded || s.Status == jobs.StatusFailed
	})
	return st
}

func TestPool_SubmitSucceeds(t *testing.T) {
	p := jobs.NewPool(2)
	id := p.Submit(context.Background(), func(ctx context.Context) error { return nil })

	st := waitTerminal(t, p, id)
	if st.Status != jobs.StatusSucceeded {
		t.Fatalf("status = %v, want succeeded", st.Status)
	}
	if st.StartedAt.IsZero() || st.FinishedAt.IsZero() {
		t.Fatalf("expected StartedAt/FinishedAt to be set: %+v", st)
	}
}

func TestPool_SubmitFails(t *testing.T) {
	p := jobs.NewPool(2)
	id := p.Submit(context.Background(), func(ctx context.Context) error {
		return errors.New("boom")
	})

	st := waitTerminal(t, p, id)
	if st.Status != jobs.StatusFailed {
		t.Fatalf("status = %v, want failed", st.Status)
	}
	if st.Err != "boom" {
		t.Fatalf("err = %q, want %q", st.Err, "boom")
	}
}

func TestPool_ConcurrencyLimit(t *testing.T) {
	const limit = 2
	p := jobs.NewPool(limit)

	var (
		mu      sync.Mutex
		current int
		maxSeen int
	)
	release := make(chan struct{})

	const n = 5
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = p.Submit(context.Background(), func(ctx context.Context) error {
			mu.Lock()
			current++
			if current > maxSeen {
				maxSeen = current
			}
			mu.Unlock()

			<-release

			mu.Lock()
			current--
			mu.Unlock()
			return nil
		})
	}

	waitUntil(t, 2*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return maxSeen == limit
	})
	close(release)

	for _, id := range ids {
		waitTerminal(t, p, id)
	}

	mu.Lock()
	defer mu.Unlock()
	if maxSeen > limit {
		t.Fatalf("maxSeen = %d, want <= %d", maxSeen, limit)
	}
}

func TestPool_ProgressReportedAndSurvivesCompletion(t *testing.T) {
	p := jobs.NewPool(1)
	id := p.Submit(context.Background(), func(ctx context.Context) error {
		progress.Report(ctx, progress.Event{Phase: "importing", Done: 1, Total: 2})
		progress.Report(ctx, progress.Event{Phase: "importing", Done: 2, Total: 2})
		return nil
	})

	st := waitTerminal(t, p, id)
	if st.Progress.Phase != "importing" || st.Progress.Done != 2 || st.Progress.Total != 2 {
		t.Fatalf("Progress = %+v, want last reported event", st.Progress)
	}
}

func TestPool_SubscribeReplaysCurrentSnapshotThenNewEvents(t *testing.T) {
	p := jobs.NewPool(1)
	gate := make(chan struct{})
	id := p.Submit(context.Background(), func(ctx context.Context) error {
		progress.Report(ctx, progress.Event{Phase: "importing", Done: 1, Total: 2})
		<-gate
		progress.Report(ctx, progress.Event{Phase: "importing", Done: 2, Total: 2})
		return nil
	})

	waitUntil(t, 2*time.Second, func() bool {
		s, _ := p.Get(id)
		return s.Progress.Done == 1
	})

	events, cancel, ok := p.Subscribe(id)
	if !ok {
		t.Fatal("Subscribe: job not found")
	}
	defer cancel()

	first := <-events
	if first.Progress.Done != 1 {
		t.Fatalf("replay snapshot Done = %d, want 1", first.Progress.Done)
	}

	close(gate)

	waitUntil(t, 2*time.Second, func() bool {
		select {
		case ev := <-events:
			return ev.Progress.Done == 2 || ev.Status == jobs.StatusSucceeded
		default:
			return false
		}
	})
}

func TestPool_SubscribeUnknownJob(t *testing.T) {
	p := jobs.NewPool(1)
	_, _, ok := p.Subscribe("does-not-exist")
	if ok {
		t.Fatal("Subscribe: expected ok=false for unknown job")
	}
}

func TestPool_GetUnknownJob(t *testing.T) {
	p := jobs.NewPool(1)
	_, ok := p.Get("does-not-exist")
	if ok {
		t.Fatal("Get: expected ok=false for unknown job")
	}
}
