package pdfparse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scainer/internal/domain"
)

func TestParseContestPDF_example10(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "examples", "10.pdf")
	f, err := os.Open(path)
	if err != nil {
		t.Skip("examples/10.pdf not found:", err)
	}
	defer f.Close()

	got, err := ParseContestPDF(f)
	if err != nil {
		t.Fatal(err)
	}
	ps, ok := got.Statements[domain.ProblemID("G")]
	if !ok {
		t.Fatalf("problem G not found, keys: %v", got.Statements)
	}
	if !strings.Contains(ps.Title, "Ася") {
		t.Errorf("title = %q", ps.Title)
	}
	if !strings.Contains(ps.RawStatement, "Асяоченьлюбит") {
		t.Errorf("rawStatement = %q", ps.RawStatement)
	}
	if !strings.Contains(ps.RawStatement, "Перваястрока") {
		t.Errorf("rawStatement should include input format: %q", ps.RawStatement)
	}
	if !strings.Contains(ps.RawStatement, "Вответекпримеру") {
		t.Errorf("rawStatement should include notes: %q", ps.RawStatement)
	}
	if len(ps.Examples) == 0 {
		t.Fatal("expected examples")
	}
	if !strings.Contains(ps.Examples[0].Input, "5") {
		t.Errorf("example input = %q", ps.Examples[0].Input)
	}
	pr, ok := got.PageRanges[domain.ProblemID("G")]
	if !ok {
		t.Fatal("page range for G")
	}
	if pr.StartPage != 1 || pr.EndPage < 1 {
		t.Errorf("page range = %+v", pr)
	}
}

func TestPagesForProblem(t *testing.T) {
	sections := []ProblemSection{{Label: "G", StartPage: 1}}
	pr, ok := pagesForProblem(2, sections, "G")
	if !ok || pr.StartPage != 1 || pr.EndPage != 2 {
		t.Fatalf("got %+v ok=%v", pr, ok)
	}
}
