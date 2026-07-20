package scoring

import (
	"testing"

	"scainer/internal/domain"
)

func pairSig(det string, contest domain.ContestID, a, b domain.ParticipantID, problem domain.ProblemID, score float64) domain.Signal {
	return domain.Signal{
		Detector: det,
		Subject: domain.NewPairSubject(contest, problem, a, b),
		Score:    score,
	}
}

func TestWeightedScore_AggregatesBySubject(t *testing.T) {
	signals := []domain.Signal{
		pairSig("always_problem", "c1", "alice", "bob", "A", 1.0),
		pairSig("always_clean", "c1", "bob", "alice", "A", 0.0), // тот же субъект: порядок участников не важен
		pairSig("always_problem", "c1", "alice", "carol", "A", 1.0),
		pairSig("always_clean", "c1", "alice", "carol", "A", 0.0),
	}

	findings := Weighted{}.Score(signals)

	if len(findings) != 2 {
		t.Fatalf("ожидали 2 находки (2 пары), получили %d", len(findings))
	}
	for _, f := range findings {
		if f.Score != 0.5 {
			t.Errorf("субъект %+v: ожидали средний score 0.5, получили %v", f.Subject, f.Score)
		}
		if len(f.Signals) != 2 {
			t.Errorf("субъект %+v: ожидали 2 сигнала, получили %d", f.Subject, len(f.Signals))
		}
	}
}

func TestWeightedScore_RankDescending(t *testing.T) {
	signals := []domain.Signal{
		pairSig("d", "c1", "a", "b", "P", 0.2),
		pairSig("d", "c1", "a", "c", "P", 0.9),
		pairSig("d", "c1", "b", "c", "P", 0.5),
	}
	findings := Weighted{}.Score(signals)
	if len(findings) != 3 {
		t.Fatalf("ожидали 3 находки, получили %d", len(findings))
	}
	for i := 1; i < len(findings); i++ {
		if findings[i-1].Score < findings[i].Score {
			t.Errorf("порядок не по убыванию score: %v < %v", findings[i-1].Score, findings[i].Score)
		}
	}
	if findings[0].Score != 0.9 {
		t.Errorf("верхняя находка должна иметь score 0.9, получили %v", findings[0].Score)
	}
}

func TestWeightedScore_DifferentContestsNotMerged(t *testing.T) {
	signals := []domain.Signal{
		pairSig("jplag", "50601", "alice", "bob", "A", 0.9),
		pairSig("jplag", "50602", "alice", "bob", "A", 0.8),
	}
	findings := Weighted{}.Score(signals)
	if len(findings) != 2 {
		t.Fatalf("AC4: ожидали 2 Finding (разный Contest), получили %d", len(findings))
	}
	keys := map[string]bool{
		domain.SubjectKey(findings[0].Subject): true,
		domain.SubjectKey(findings[1].Subject): true,
	}
	if !keys["pair|50601|A|alice|bob"] || !keys["pair|50602|A|alice|bob"] {
		t.Fatalf("ключи: %v", keys)
	}
}
