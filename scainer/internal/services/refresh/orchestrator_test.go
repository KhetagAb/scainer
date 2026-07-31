package refresh_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements"
	"scainer/internal/services/statements/explain"
	"scainer/pkg/store"
)

func TestRefreshContestStatements(t *testing.T) {
	dir := t.TempDir()
	store := statements.NewProblemStore(dir)
	reg := newFakeRegistry()
	reg.byID["50506"] = contests.ContestRecord{
		Contest: contests.Contest{ID: "50506", ParallelID: "10"},
	}

	pdfPath := filepath.Join("..", "..", "..", "examples", "10.pdf")
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Skip("examples/10.pdf:", err)
	}

	svc := statements.NewService(reg, map[string]statements.Provider{
		statements.SourceLksh: &fileProvider{data: data},
	}, store, explain.NewSubmissionLabelResolver(store.NewMem()))

	if err := svc.RefreshContestStatements(context.Background(), "50506"); err != nil {
		t.Fatal(err)
	}
	if !store.HasSourcePDF("50506") {
		t.Fatal("expected source.pdf")
	}
	f, err := store.OpenStatementPDF("50506", "G")
	if err != nil {
		t.Fatalf("open G.pdf: %v", err)
	}
	f.Close()
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

type fileProvider struct {
	data []byte
}

func (fileProvider) SourceType() string { return statements.SourceLksh }

func (p *fileProvider) Fetch(_ context.Context, _ statements.Contest) (statements.Document, error) {
	return statements.Document{
		Body:        io.NopCloser(bytes.NewReader(p.data)),
		ContentType: "application/pdf",
		Filename:    "10.pdf",
	}, nil
}
