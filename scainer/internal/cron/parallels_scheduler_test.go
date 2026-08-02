package cron

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/importer"
	"scainer/pkg/ejudge/servecontrol"
)

func TestValidateParallelsSchedule(t *testing.T) {
	if err := ValidateParallelsSchedule("0 9 * * *"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateParallelsSchedule(""); err != nil {
		t.Fatal(err)
	}
	if err := ValidateParallelsSchedule("bad cron"); err == nil {
		t.Fatal("expected error for bad schedule")
	}
}

func TestParallelsSchedulerTick(t *testing.T) {
	importer.Register("ejudge", func(*yaml.Node) (importer.Importer, error) {
		return stubImporter{}, nil
	})

	svc := contests.NewService(parallelsFakeRegistry{}, fakeAnalysisRepo{}, "ejudge")
	var calls int
	list := func(context.Context) ([]servecontrol.Brief, error) {
		calls++
		return []servecontrol.Brief{
			{ID: 50501, Name: "ЛКШ.2026.Параллель 5.День 01.A"},
			{ID: 50502, Name: "ЛКШ.2026.Параллель 5.Template"},
		}, nil
	}

	sched := NewParallelsScheduler(svc, list, "cron-user")
	sched.tick(context.Background())

	if calls != 1 {
		t.Fatalf("list calls=%d", calls)
	}
}

type parallelsFakeRegistry struct{}

func (parallelsFakeRegistry) Put(context.Context, contests.ContestRecord) error { return nil }
func (parallelsFakeRegistry) Get(context.Context, domain.ContestID) (contests.ContestRecord, bool, error) {
	return contests.ContestRecord{}, false, nil
}
func (parallelsFakeRegistry) Delete(context.Context, domain.ContestID) error { return nil }
func (parallelsFakeRegistry) List(context.Context) ([]contests.ContestRecord, error) {
	return nil, nil
}

type fakeAnalysisRepo struct{}

func (fakeAnalysisRepo) Put(context.Context, contests.AnalysisSnapshot) error { return nil }
func (fakeAnalysisRepo) Get(context.Context, domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	return contests.AnalysisSnapshot{}, false, nil
}
func (fakeAnalysisRepo) Delete(context.Context, domain.ContestID) error { return nil }

type stubImporter struct{}

func (stubImporter) Name() string { return "ejudge" }
func (stubImporter) Import(context.Context, importer.Store) (importer.Result, error) {
	return importer.Result{}, nil
}
