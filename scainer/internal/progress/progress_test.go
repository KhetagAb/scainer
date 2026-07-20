package progress_test

import (
	"context"
	"testing"

	"scainer/internal/progress"
)

func TestReport_NoReporter_NoPanic(t *testing.T) {
	progress.Report(context.Background(), progress.Event{Phase: "importing", Done: 1, Total: 2})
}

func TestReport_WithReporter_Delivered(t *testing.T) {
	var got []progress.Event
	ctx := progress.With(context.Background(), func(e progress.Event) {
		got = append(got, e)
	})

	progress.Report(ctx, progress.Event{Phase: "importing", Done: 1, Total: 3})
	progress.Report(ctx, progress.Event{Phase: "importing", Done: 2, Total: 3})

	if len(got) != 2 {
		t.Fatalf("got %d events, want 2", len(got))
	}
	if got[0].Done != 1 || got[1].Done != 2 {
		t.Fatalf("unexpected events: %+v", got)
	}
}

func TestWith_NilReporter_NoOp(t *testing.T) {
	ctx := progress.With(context.Background(), nil)
	progress.Report(ctx, progress.Event{Phase: "importing"})
}
