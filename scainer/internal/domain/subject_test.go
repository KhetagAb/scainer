package domain_test

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/lksh/scainer/internal/domain"
)

func TestSubjectKey_PairOrderIndependent(t *testing.T) {
	a := domain.NewPairSubject("c", "A", "alice", "bob")
	b := domain.NewPairSubject("c", "A", "bob", "alice")
	if domain.SubjectKey(a) != domain.SubjectKey(b) {
		t.Fatalf("%q != %q", domain.SubjectKey(a), domain.SubjectKey(b))
	}
}

func TestSubjectKey_ContestSeparates(t *testing.T) {
	a := domain.NewPairSubject("1", "A", "alice", "bob")
	b := domain.NewPairSubject("2", "A", "alice", "bob")
	if domain.SubjectKey(a) == domain.SubjectKey(b) {
		t.Fatal("разный Contest должен давать разный ключ")
	}
}

func TestSubject_Kind(t *testing.T) {
	var s domain.Subject = domain.NewPairSubject("c", "A", "x", "y")
	if s.Kind() != domain.SubjectPair {
		t.Fatalf("Kind = %q", s.Kind())
	}
}

func TestSubmissionSubject_KeyAndJSON(t *testing.T) {
	s := domain.NewSubmissionSubject("50601", "ejudge:50601:1")
	if got := domain.SubjectKey(s); got != "submission|50601|ejudge:50601:1" {
		t.Fatalf("key = %q", got)
	}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["kind"] != "submission" || m["contest"] != "50601" || m["submission"] != "ejudge:50601:1" {
		t.Fatalf("got %#v", m)
	}
}

func TestPairSubject_MarshalJSON(t *testing.T) {
	raw, err := json.Marshal(domain.NewPairSubject("50601", "A", "alice", "bob"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["kind"] != "pair" || m["contest"] != "50601" || m["problem"] != "A" || m["a"] != "alice" || m["b"] != "bob" {
		t.Fatalf("got %#v", m)
	}
}

func findingFixture(subj domain.Subject) domain.Finding {
	return domain.Finding{
		Subject: subj,
		Score:   0.75,
		Signals: []domain.Signal{
			{
				Detector: "jplag",
				AI:       false,
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
				Meta: map[string]any{"run_id": "1"},
			},
		},
	}
}

func TestFinding_BSONRoundTrip(t *testing.T) {
	subjects := []domain.Subject{
		domain.NewSubmissionSubject("50601", "ejudge:50601:1"),
		domain.NewPairSubject("50601", "A", "alice", "bob"),
		domain.NewParticipantProblemSubject("50601", "A", "alice"),
		domain.NewParticipantSubject("alice"),
	}
	for _, subj := range subjects {
		t.Run(string(subj.Kind()), func(t *testing.T) {
			want := findingFixture(subj)

			raw, err := bson.Marshal(want)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			var got domain.Finding
			if err := bson.Unmarshal(raw, &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if got.Subject.Kind() != subj.Kind() || domain.SubjectKey(got.Subject) != domain.SubjectKey(subj) {
				t.Fatalf("Finding.Subject: got %#v want %#v", got.Subject, subj)
			}
			if got.Score != want.Score {
				t.Fatalf("Finding.Score: got %v want %v", got.Score, want.Score)
			}
			if len(got.Signals) != 1 {
				t.Fatalf("Signals: got %d want 1", len(got.Signals))
			}
			sig := got.Signals[0]
			if sig.Detector != "jplag" || sig.Score != 0.9 {
				t.Fatalf("Signal: got %#v", sig)
			}
			if sig.Subject.Kind() != subj.Kind() || domain.SubjectKey(sig.Subject) != domain.SubjectKey(subj) {
				t.Fatalf("Signal.Subject: got %#v want %#v", sig.Subject, subj)
			}
			if len(sig.Evidence) != 1 || len(sig.Evidence[0].Spans) != 1 {
				t.Fatalf("Evidence/Spans: got %#v", sig.Evidence)
			}
			if sig.Evidence[0].Spans[0].Submission != "ejudge:1:1" || sig.Evidence[0].Spans[0].EndLine != 5 {
				t.Fatalf("Span: got %#v", sig.Evidence[0].Spans[0])
			}
		})
	}
}
