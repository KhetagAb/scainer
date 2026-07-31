package statements_test

import (
	"context"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements"
	"scainer/internal/services/statements/explain"
	"scainer/pkg/store"
)

func TestResolveStatementSource(t *testing.T) {
	t.Parallel()
	if got := statements.ResolveStatementSource(statements.Contest{ParallelID: "5"}); got != statements.SourceLksh {
		t.Fatalf("got %q want %q", got, statements.SourceLksh)
	}
	if got := statements.ResolveStatementSource(statements.Contest{}); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestServiceFetch_noParallel(t *testing.T) {
	reg := newFakeRegistry()
	reg.byID["50152"] = contests.ContestRecord{
		Contest: contests.Contest{ID: "50152"},
		Source:  contests.SourceSpec{Type: "ejudge"},
	}
	svc := statements.NewService(reg, map[string]statements.Provider{
		statements.SourceLksh: stubProvider{t: t},
	}, nil, explain.NewSubmissionLabelResolver(store.NewMem()))
	_, err := svc.Fetch(context.Background(), domain.ContestID("50152"))
	if err != statements.ErrNotAvailable {
		t.Fatalf("err = %v want ErrNotAvailable", err)
	}
}

type fakeRegistry struct {
	byID map[domain.ContestID]contests.ContestRecord
}

func newFakeRegistry() *fakeRegistry {
	return &fakeRegistry{byID: make(map[domain.ContestID]contests.ContestRecord)}
}

func (r *fakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	rec, ok := r.byID[id]
	return rec, ok, nil
}

type stubProvider struct {
	t *testing.T
}

func (stubProvider) SourceType() string { return statements.SourceLksh }

func (p stubProvider) Fetch(context.Context, statements.Contest) (statements.Document, error) {
	p.t.Fatal("Fetch should not be called")
	return statements.Document{}, nil
}
