package domain

import "testing"

func TestParseVerdict(t *testing.T) {
	if ParseVerdict("OK") != VerdictOK || ParseVerdict("WA") != VerdictWA || ParseVerdict("RJ") != VerdictRJ || ParseVerdict("PR") != VerdictPR {
		t.Fatal("known")
	}
	if ParseVerdict("CE") != VerdictUnknown || ParseVerdict("RT") != VerdictUnknown || ParseVerdict("PD") != VerdictUnknown || ParseVerdict("") != VerdictUnknown {
		t.Fatal("unknown")
	}
}
