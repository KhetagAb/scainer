package ejudge

import "testing"

func TestVerdictFromStatus(t *testing.T) {
	if VerdictFromStatus(0) != VerdictOK || VerdictFromStatus(5) != VerdictWA || VerdictFromStatus(12) != VerdictML {
		t.Fatal("codes")
	}
	if VerdictFromStatus(16) != VerdictPR {
		t.Fatal("PR")
	}
	if VerdictFromStatus(17) != VerdictRJ {
		t.Fatal("RJ")
	}
	if VerdictFromStatus(14) != VerdictRJ {
		t.Fatal("SV -> RJ")
	}
	if VerdictFromStatus(1) != VerdictCE {
		t.Fatal("CE -> CE")
	}
	if VerdictFromStatus(2) != VerdictUnknown { // RT
		t.Fatal("RT -> unknown")
	}
	if VerdictFromStatus(11) != VerdictUnknown { // PD
		t.Fatal("PD -> unknown")
	}
	if VerdictFromStatus(8) != VerdictUnknown { // AC (accepted for testing)
		t.Fatal("AC -> unknown")
	}
}

func TestParseVerdict(t *testing.T) {
	if ParseVerdict("SV") != VerdictRJ {
		t.Fatal("SV -> RJ")
	}
	if ParseVerdict("CE") != VerdictCE {
		t.Fatal("CE -> CE")
	}
	if ParseVerdict("AC") != VerdictUnknown {
		t.Fatal("AC -> unknown")
	}
}

func TestStatusCode(t *testing.T) {
	code, ok := StatusCode(VerdictRJ)
	if !ok || code != 17 {
		t.Fatalf("RJ: %d %v", code, ok)
	}
	if _, ok := StatusCode(VerdictUnknown); ok {
		t.Fatal("unknown must fail")
	}
}
