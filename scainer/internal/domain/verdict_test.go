package domain

import "testing"

func TestParseVerdict(t *testing.T) {
	if ParseVerdict("OK") != VerdictOK || ParseVerdict("WA") != VerdictWA {
		t.Fatal("known")
	}
	if ParseVerdict("CE") != VerdictUnknown || ParseVerdict("RT") != VerdictUnknown || ParseVerdict("") != VerdictUnknown {
		t.Fatal("unknown")
	}
}

func TestVerdictFromEjudgeStatus(t *testing.T) {
	if VerdictFromEjudgeStatus(0) != VerdictOK || VerdictFromEjudgeStatus(5) != VerdictWA || VerdictFromEjudgeStatus(12) != VerdictML {
		t.Fatal("codes")
	}
	if VerdictFromEjudgeStatus(1) != VerdictUnknown { // CE
		t.Fatal("CE -> unknown")
	}
}
