package contests_test

import (
	"context"
	"errors"
	"testing"

	"gopkg.in/yaml.v3"

	"scainer/internal/services/contests"
	"scainer/internal/services/importer"
	"scainer/pkg/ejudge/servecontrol"
)

func init() {
	importer.Register("ejudge", func(*yaml.Node) (importer.Importer, error) {
		return stubImporter{}, nil
	})
}

func TestClassifyEjudgeContests(t *testing.T) {
	briefs := []servecontrol.Brief{
		{ID: 50050, Name: "ЛКШ.2026.Параллель R.Template"},
		{ID: 50051, Name: "ЛКШ.2026.Параллель R.День 01.Разнобой"},
		{ID: 50201, Name: "ЛКШ.2026.Август.Параллель 2.День 01.STL"},
		{ID: 50999, Name: "ЛКШ.2026.Олимпиада Div.1"},
	}

	toImport, skipped := contests.ClassifyEjudgeContests(briefs)
	if len(toImport) != 1 || toImport[0].ID != "50051" || toImport[0].ParallelID != "R" {
		t.Fatalf("toImport: %+v", toImport)
	}
	if len(skipped) != 3 {
		t.Fatalf("skipped: %+v", skipped)
	}
	if skipped[0].Reason != "шаблон" {
		t.Fatalf("template reason: %q", skipped[0].Reason)
	}
	if skipped[1].Reason != "параллель не распознана" {
		t.Fatalf("parallel reason: %q", skipped[1].Reason)
	}
}

func TestImportFromEjudge_AddsAndDuplicates(t *testing.T) {
	ctx := context.Background()
	svc := contests.NewService(newFakeRegistry(), newFakeAnalysisRepository(), "ejudge")

	list := func(context.Context) ([]servecontrol.Brief, error) {
		return []servecontrol.Brief{
			{ID: 50501, Name: "ЛКШ.2026.Параллель 5.День 01.A"},
			{ID: 50502, Name: "ЛКШ.2026.Параллель 5.День 02.B"},
			{ID: 50503, Name: "ЛКШ.2026.Параллель 5.Template"},
		}, nil
	}

	res, err := svc.ImportFromEjudge(ctx, list)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 2 || len(res.Skipped) != 1 || len(res.Duplicates) != 0 {
		t.Fatalf("first import: added=%d skipped=%d duplicates=%d", len(res.Added), len(res.Skipped), len(res.Duplicates))
	}

	res, err = svc.ImportFromEjudge(ctx, list)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 0 || len(res.Duplicates) != 2 {
		t.Fatalf("second import: added=%d duplicates=%d", len(res.Added), len(res.Duplicates))
	}
}

func TestImportFromEjudge_ListError(t *testing.T) {
	ctx := context.Background()
	svc := contests.NewService(newFakeRegistry(), newFakeAnalysisRepository(), "ejudge")

	_, err := svc.ImportFromEjudge(ctx, func(context.Context) ([]servecontrol.Brief, error) {
		return nil, errors.New("ejudge down")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
