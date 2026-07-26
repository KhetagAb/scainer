package repository_test

import (
	"context"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/repository"
	"scainer/internal/services/contests"
)

func signalsFixture(contest domain.ContestID) contests.Signals {
	subj := domain.NewPairSubject(contest, "A", "alice", "bob")
	return contests.Signals{
		"jplag": {
			"problem:A:cpp": {
				{
					Detector: "jplag",
					Subject:  subj,
					Score:    0.9,
					Evidence: []domain.Evidence{
						{
							Kind:        "jplag_match",
							Description: "similarity",
							Spans: []domain.Span{
								{Submission: "ejudge:1:1", StartLine: 1, EndLine: 5},
							},
						},
					},
				},
			},
		},
	}
}

func TestAnalysisRepository_PutGetDelete(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()
	repo := repository.NewAnalysisRepository(db)

	const contest domain.ContestID = "contest01"

	if _, ok, err := repo.Get(ctx, contest); err != nil || ok {
		t.Fatalf("Get до Put: ok=%v err=%v (ожидали ok=false, err=nil)", ok, err)
	}

	snap := contests.AnalysisSnapshot{
		ContestID:  contest,
		Signals:    signalsFixture(contest),
		ComputedAt: time.Now().UTC().Truncate(time.Millisecond),
	}
	if err := repo.Put(ctx, snap); err != nil {
		t.Fatalf("Put: %v", err)
	}

	got, ok, err := repo.Get(ctx, contest)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !ok {
		t.Fatal("ожидали ok=true после Put")
	}
	sigs := got.Signals["jplag"]["problem:A:cpp"]
	if len(sigs) != 1 {
		t.Fatalf("Signals: got %d want 1", len(sigs))
	}
	if sigs[0].Detector != "jplag" {
		t.Fatalf("Detector: got %q want jplag", sigs[0].Detector)
	}
	if len(sigs[0].Evidence) != 1 || len(sigs[0].Evidence[0].Spans) != 1 {
		t.Fatalf("Evidence/Spans: %+v", sigs[0].Evidence)
	}
	if sigs[0].Evidence[0].Spans[0].EndLine != 5 {
		t.Fatalf("Span.EndLine: got %d want 5", sigs[0].Evidence[0].Spans[0].EndLine)
	}

	if err := repo.Delete(ctx, contest); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, err := repo.Get(ctx, contest); err != nil || ok {
		t.Fatalf("Get после Delete: ok=%v err=%v (ожидали ok=false)", ok, err)
	}
}

func TestAnalysisRepository_PutOverwrites(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()
	repo := repository.NewAnalysisRepository(db)
	const contest domain.ContestID = "contest01"

	if err := repo.Put(ctx, contests.AnalysisSnapshot{ContestID: contest, Signals: signalsFixture(contest)}); err != nil {
		t.Fatalf("first Put: %v", err)
	}
	if err := repo.Put(ctx, contests.AnalysisSnapshot{ContestID: contest, Signals: nil}); err != nil {
		t.Fatalf("second Put: %v", err)
	}

	got, ok, err := repo.Get(ctx, contest)
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if len(got.Signals) != 0 {
		t.Fatalf("второй Put должен был полностью заменить документ: %+v", got.Signals)
	}
}
