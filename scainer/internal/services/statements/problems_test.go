package statements_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements"
	"scainer/internal/services/statements/explain"
	substore "scainer/pkg/store"
)

func TestIngestAndGetProblemStatement(t *testing.T) {
	dir := t.TempDir()
	store := statements.NewProblemStore(dir)
	reg := newFakeRegistry()
	reg.byID["50506"] = contests.ContestRecord{
		Contest: contests.Contest{ID: "50506"},
	}
	svc := statements.NewService(reg, nil, store, explain.NewSubmissionLabelResolver(substore.NewMem()))

	pdfPath := filepath.Join("..", "..", "..", "examples", "10.pdf")
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Skip("examples/10.pdf:", err)
	}
	defer f.Close()

	if err := svc.IngestContestPDF(context.Background(), domain.ContestID("50506"), f); err != nil {
		t.Fatal(err)
	}

	ps, err := svc.GetProblemStatement(context.Background(), "50506", "G")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ps.Statement, "Асяоченьлюбит") {
		t.Errorf("statement: %q", ps.Statement)
	}

	body, err := svc.OpenProblemStatementPDF(context.Background(), "50506", "G")
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	data, err := os.ReadFile(filepath.Join(dir, "problem-statements", "50506", "G.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 100 {
		t.Fatalf("cropped pdf too small: %d bytes", len(data))
	}
}
