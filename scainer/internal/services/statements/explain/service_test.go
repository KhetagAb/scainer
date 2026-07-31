package explain_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements"
)

type fakeLabelResolver struct {
	labels map[domain.ProblemID]string
}

func (r fakeLabelResolver) ProblemLabel(_ context.Context, _ domain.ContestID, problemID domain.ProblemID) (string, bool, error) {
	label, ok := r.labels[problemID]
	return label, ok, nil
}

type fakeRegistry struct {
	byID map[domain.ContestID]contests.ContestRecord
}

func (r *fakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	rec, ok := r.byID[id]
	return rec, ok, nil
}

func TestExplain_resolvesSlugToLetter(t *testing.T) {
	dir := t.TempDir()
	store := statements.NewProblemStore(dir)
	reg := &fakeRegistry{byID: map[domain.ContestID]contests.ContestRecord{
		"50506": {Contest: contests.Contest{ID: "50506"}},
	}}
	labels := fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}
	stmts := statements.NewService(reg, nil, store, labels)

	pdfPath := filepath.Join("..", "..", "..", "..", "examples", "10.pdf")
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Skip("examples/10.pdf:", err)
	}
	defer f.Close()

	if err := stmts.IngestContestPDF(context.Background(), domain.ContestID("50506"), f); err != nil {
		t.Fatal(err)
	}

	ex, err := stmts.Explain.Explain(context.Background(), "50506", "gray-code")
	if err != nil {
		t.Fatal(err)
	}
	if ex.Problem != "G" {
		t.Fatalf("problem=%q want G", ex.Problem)
	}
	if ex.Title == "" {
		t.Fatal("expected non-empty title")
	}
	if !strings.Contains(ex.Statement, "Асяоченьлюбит") {
		t.Errorf("statement: %q", ex.Statement)
	}
}
