package repository_test

import (
	"context"
	"testing"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/repository"
	"scainer/internal/services/contests"
)

func TestContestRepository_PutGetListDelete(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()
	repo := repository.NewContestRepository(db)

	var cfg yaml.Node
	if err := yaml.Unmarshal([]byte("contest_id: 50501\n"), &cfg); err != nil {
		t.Fatal(err)
	}

	rec := contests.ContestRecord{
		Contest: contests.Contest{
			ID:         "contest01",
			Name:       "Тестовый контест",
			ParallelID: "par1",
			Problems: []domain.ContestProblem{
				{ID: "A", Name: "A"},
				{ID: "B", Name: "B"},
			},
		},
		Source: contests.SourceSpec{Type: "ejudge", Config: cfg},
	}
	if err := repo.Put(ctx, rec); err != nil {
		t.Fatalf("Put: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ожидали 1 запись, получили %d", len(list))
	}
	got := list[0]
	if got.Contest.ParallelID != "par1" {
		t.Fatalf("Contest: %+v", got.Contest)
	}
	if len(got.Contest.Problems) != 2 || got.Contest.Problems[0].ID != "A" {
		t.Fatalf("Problems: %+v", got.Contest.Problems)
	}
	if got.Source.Type != "ejudge" {
		t.Fatalf("Source.Type: %q", got.Source.Type)
	}
	var roundTripped struct {
		ContestID int `yaml:"contest_id"`
	}
	if err := got.Source.Config.Decode(&roundTripped); err != nil {
		t.Fatalf("decode Source.Config: %v", err)
	}
	if roundTripped.ContestID != 50501 {
		t.Fatalf("Source.Config.contest_id: got %d want 50501", roundTripped.ContestID)
	}

	if err := repo.Put(ctx, rec); err != nil {
		t.Fatalf("Put (update): %v", err)
	}
	list, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if err := repo.Delete(ctx, "contest01"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	list, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("ожидали 0 записей после удаления, получили %d", len(list))
	}

	if err := repo.Delete(ctx, "missing"); err != nil {
		t.Fatalf("Delete отсутствующей записи не должен ошибаться: %v", err)
	}
}

func TestContestRepository_PersistsAcrossProcesses(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()

	repo1 := repository.NewContestRepository(db)
	if err := repo1.Put(ctx, contests.ContestRecord{
		Contest:   contests.Contest{ID: "contest01", ParallelID: "par1"},
		Source: contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Put: %v", err)
	}

	repo2 := repository.NewContestRepository(db)
	list, err := repo2.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Contest.ParallelID != "par1" {
		t.Fatalf("не пережило рестарт: %+v", list)
	}
}
