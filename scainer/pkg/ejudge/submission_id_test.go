package ejudge_test

import (
	"testing"

	"scainer/pkg/ejudge"
)

func TestSubmissionKey_RoundTrip(t *testing.T) {
	k := ejudge.SubmissionKey{ContestID: 50501, RunID: 7}
	if got := k.ID(); got != "ejudge:50501:7" {
		t.Fatalf("ID=%q", got)
	}
	if k.Contest() != "50501" {
		t.Fatalf("Contest=%q", k.Contest())
	}
	parsed, err := ejudge.ParseSubmissionID(k.ID())
	if err != nil || parsed != k {
		t.Fatalf("parse: %+v %v", parsed, err)
	}
	if _, err := ejudge.ParseSubmissionID("folder:x"); err == nil {
		t.Fatal("expected error")
	}
}
