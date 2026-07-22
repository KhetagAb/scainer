package ejudge

import (
	"testing"

	"scainer/internal/domain"
)

func TestVerdictFromStatus(t *testing.T) {
	if VerdictFromStatus(0) != domain.VerdictOK || VerdictFromStatus(5) != domain.VerdictWA || VerdictFromStatus(12) != domain.VerdictML {
		t.Fatal("codes")
	}
	if VerdictFromStatus(16) != domain.VerdictPR {
		t.Fatal("PR")
	}
	if VerdictFromStatus(17) != domain.VerdictRJ {
		t.Fatal("RJ")
	}
	if VerdictFromStatus(1) != domain.VerdictUnknown { // CE
		t.Fatal("CE -> unknown")
	}
	if VerdictFromStatus(11) != domain.VerdictUnknown { // PD (pending, not review)
		t.Fatal("PD -> unknown")
	}
}

func TestStatusCode(t *testing.T) {
	code, ok := StatusCode(domain.VerdictRJ)
	if !ok || code != 17 {
		t.Fatalf("RJ: %d %v", code, ok)
	}
	if _, ok := StatusCode(domain.VerdictUnknown); ok {
		t.Fatal("unknown must fail")
	}
}
