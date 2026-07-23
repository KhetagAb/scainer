package repository_test

import (
	"context"
	"testing"
	"time"

	"scainer/internal/services/contests"
	"scainer/internal/domain"
	"scainer/internal/repository"
)

func findingsFixture(contest domain.ContestID) []domain.Finding {
	return []domain.Finding{
		{
			Subject: domain.NewPairSubject(contest, "A", "alice", "bob"),
			Score:   0.9,
			Signals: []domain.Signal{
				{
					Detector: "jplag",
					Subject:  domain.NewPairSubject(contest, "A", "alice", "bob"),
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

func TestFindingsRepository_PutGetDelete(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()
	repo := repository.NewFindingsRepository(db)

	const contest domain.ContestID = "contest01"

	if _, ok, err := repo.Get(ctx, contest); err != nil || ok {
		t.Fatalf("Get до Put: ok=%v err=%v (ожидали ok=false, err=nil)", ok, err)
	}

	snap := contests.FindingsSnapshot{
		ContestID:  contest,
		Findings:   findingsFixture(contest),
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
	if len(got.Findings) != 1 {
		t.Fatalf("Findings: got %d want 1", len(got.Findings))
	}
	f := got.Findings[0]
	if f.Subject.Kind() != domain.SubjectPair || domain.SubjectKey(f.Subject) != domain.SubjectKey(snap.Findings[0].Subject) {
		t.Fatalf("Subject не пережил round-trip: %#v", f.Subject)
	}
	if len(f.Signals) != 1 || f.Signals[0].Detector != "jplag" {
		t.Fatalf("Signals: %+v", f.Signals)
	}
	if len(f.Signals[0].Evidence) != 1 || len(f.Signals[0].Evidence[0].Spans) != 1 {
		t.Fatalf("Evidence/Spans: %+v", f.Signals[0].Evidence)
	}
	if f.Signals[0].Evidence[0].Spans[0].EndLine != 5 {
		t.Fatalf("Span.EndLine: got %d want 5", f.Signals[0].Evidence[0].Spans[0].EndLine)
	}

	if err := repo.Delete(ctx, contest); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok, err := repo.Get(ctx, contest); err != nil || ok {
		t.Fatalf("Get после Delete: ok=%v err=%v (ожидали ok=false)", ok, err)
	}
}

func TestFindingsRepository_PutOverwrites(t *testing.T) {
	db := mustDB(t)
	ctx := context.Background()
	repo := repository.NewFindingsRepository(db)
	const contest domain.ContestID = "contest01"

	if err := repo.Put(ctx, contests.FindingsSnapshot{ContestID: contest, Findings: findingsFixture(contest)}); err != nil {
		t.Fatalf("first Put: %v", err)
	}
	if err := repo.Put(ctx, contests.FindingsSnapshot{ContestID: contest, Findings: nil}); err != nil {
		t.Fatalf("second Put: %v", err)
	}

	got, ok, err := repo.Get(ctx, contest)
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if len(got.Findings) != 0 {
		t.Fatalf("второй Put должен был полностью заменить документ: %+v", got.Findings)
	}
}
