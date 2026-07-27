package analyze

import (
	"context"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/ejudge"
	"scainer/pkg/store"
)

type runnerFakeRegistry struct {
	byID map[domain.ContestID]contests.ContestRecord
}

func (r *runnerFakeRegistry) Put(_ context.Context, rec contests.ContestRecord) error {
	r.byID[rec.Contest.ID] = rec
	return nil
}

func (r *runnerFakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	rec, ok := r.byID[id]
	return rec, ok, nil
}

func (r *runnerFakeRegistry) Delete(_ context.Context, id domain.ContestID) error {
	delete(r.byID, id)
	return nil
}

func (r *runnerFakeRegistry) List(context.Context) ([]contests.ContestRecord, error) {
	out := make([]contests.ContestRecord, 0, len(r.byID))
	for _, rec := range r.byID {
		out = append(out, rec)
	}
	return out, nil
}

type runnerFakeAnalysisRepository struct {
	byID map[domain.ContestID]contests.AnalysisSnapshot
}

func (f *runnerFakeAnalysisRepository) Put(_ context.Context, snap contests.AnalysisSnapshot) error {
	f.byID[snap.ContestID] = snap
	return nil
}

func (f *runnerFakeAnalysisRepository) Get(_ context.Context, id domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	snap, ok := f.byID[id]
	return snap, ok, nil
}

func (f *runnerFakeAnalysisRepository) Delete(_ context.Context, id domain.ContestID) error {
	delete(f.byID, id)
	return nil
}

func TestRunner_ResetContestData(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	reg := &runnerFakeRegistry{byID: make(map[domain.ContestID]contests.ContestRecord)}
	repo := &runnerFakeAnalysisRepository{byID: make(map[domain.ContestID]contests.AnalysisSnapshot)}
	runner := NewRunner(reg, st, repo, nil)

	cfgRaw, err := yaml.Marshal(map[string]int{"contest_id": 50501})
	if err != nil {
		t.Fatal(err)
	}
	var cfgNode yaml.Node
	if err := yaml.Unmarshal(cfgRaw, &cfgNode); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	reg.byID["50501"] = contests.ContestRecord{
		Contest: contests.Contest{
			ID:             "50501",
			LastImportedAt: &now,
		},
		Source: contests.SourceSpec{Type: "ejudge", Config: *cfgNode.Content[0]},
	}
	if err := st.Put(ctx, []domain.Submission{
		{ID: "s1", Contest: "50501", Problem: "A", Participant: "p"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.SetCursor(ctx, ejudge.ImportCursorKey(50501), "10"); err != nil {
		t.Fatal(err)
	}
	repo.byID["50501"] = contests.AnalysisSnapshot{
		ContestID:  "50501",
		ComputedAt: now,
	}

	if err := runner.ResetContestData(ctx, "50501"); err != nil {
		t.Fatalf("ResetContestData: %v", err)
	}

	byProblem, err := st.ByProblem(ctx, "50501")
	if err != nil {
		t.Fatal(err)
	}
	if len(byProblem) != 0 {
		t.Fatalf("submissions remain: %+v", byProblem)
	}
	if _, ok, err := repo.Get(ctx, "50501"); err != nil || ok {
		t.Fatalf("analysis snapshot should be deleted: ok=%v err=%v", ok, err)
	}
	if _, ok, _ := st.GetCursor(ctx, ejudge.ImportCursorKey(50501)); ok {
		t.Fatal("ejudge cursor should be deleted")
	}
	rec, ok, err := reg.Get(ctx, "50501")
	if err != nil || !ok {
		t.Fatalf("contest record: ok=%v err=%v", ok, err)
	}
	if rec.Contest.LastImportedAt != nil {
		t.Fatal("lastImportedAt should be cleared")
	}
}
