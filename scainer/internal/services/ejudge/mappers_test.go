package ejudge

import (
	"testing"

	"scainer/internal/domain"
)

func TestVerdictFromEjudgeStatus(t *testing.T) {
	if verdictFromEjudgeStatus(0) != domain.VerdictOK || verdictFromEjudgeStatus(5) != domain.VerdictWA || verdictFromEjudgeStatus(12) != domain.VerdictML {
		t.Fatal("known codes")
	}
	if verdictFromEjudgeStatus(16) != domain.VerdictPR {
		t.Fatal("PR")
	}
	if verdictFromEjudgeStatus(17) != domain.VerdictRJ {
		t.Fatal("RJ")
	}
	if verdictFromEjudgeStatus(14) != domain.VerdictRJ {
		t.Fatal("SV code")
	}
	if verdictFromEjudgeStatus(1) != domain.VerdictCE {
		t.Fatal("CE")
	}
	if verdictFromEjudgeStatus(2) != domain.VerdictRT {
		t.Fatal("RT")
	}
	if verdictFromEjudgeStatus(4) != domain.VerdictPE {
		t.Fatal("PE")
	}
	if verdictFromEjudgeStatus(7) != domain.VerdictWA {
		t.Fatal("PT code")
	}
	if verdictFromEjudgeStatus(9) != domain.VerdictIG {
		t.Fatal("IG")
	}
	if verdictFromEjudgeStatus(15) != domain.VerdictRU {
		t.Fatal("WT code")
	}
	if verdictFromEjudgeStatus(23) != domain.VerdictPR {
		t.Fatal("SM code")
	}
	for _, code := range []int{6, 20, 21, 22, 95, 96, 97, 98, 99} {
		if verdictFromEjudgeStatus(code) != domain.VerdictRU {
			t.Fatalf("running code %d", code)
		}
	}
	if verdictFromEjudgeStatus(11) != domain.VerdictUnknown { // PD
		t.Fatal("PD")
	}
	if verdictFromEjudgeStatus(8) != domain.VerdictUnknown { // AC
		t.Fatal("AC")
	}
}

func TestParseEjudgeStatusStr(t *testing.T) {
	if parseEjudgeStatusStr("OK") != domain.VerdictOK || parseEjudgeStatusStr("WA") != domain.VerdictWA ||
		parseEjudgeStatusStr("RJ") != domain.VerdictRJ || parseEjudgeStatusStr("PR") != domain.VerdictPR {
		t.Fatal("known")
	}
	if parseEjudgeStatusStr("CE") != domain.VerdictCE || parseEjudgeStatusStr("RT") != domain.VerdictRT ||
		parseEjudgeStatusStr("IG") != domain.VerdictIG || parseEjudgeStatusStr("PE") != domain.VerdictPE {
		t.Fatal("RT/IG/PE")
	}
	if parseEjudgeStatusStr("SV") != domain.VerdictRJ {
		t.Fatal("SV -> RJ")
	}
	if parseEjudgeStatusStr("PT") != domain.VerdictWA {
		t.Fatal("PT -> WA")
	}
	for _, s := range []string{"CG", "CD", "AV", "VS", "VT", "EM", "RU", "WT", "CF"} {
		if parseEjudgeStatusStr(s) != domain.VerdictRU {
			t.Fatalf("%s -> RU", s)
		}
	}
	if parseEjudgeStatusStr("SM") != domain.VerdictPR {
		t.Fatal("SM -> PR")
	}
	if parseEjudgeStatusStr("AC") != domain.VerdictUnknown || parseEjudgeStatusStr("PD") != domain.VerdictUnknown ||
		parseEjudgeStatusStr("") != domain.VerdictUnknown {
		t.Fatal("unknown")
	}
}

func TestMapToDomainVerdict(t *testing.T) {
	if mapToDomainVerdict(17, "RJ") != domain.VerdictRJ {
		t.Fatal("status_str wins")
	}
	if mapToDomainVerdict(0, "") != domain.VerdictOK {
		t.Fatal("fallback to status code")
	}
}

func TestDomainVerdictToStatusCode(t *testing.T) {
	code, ok := domainVerdictToStatusCode(domain.VerdictRJ)
	if !ok || code != 17 {
		t.Fatalf("RJ: code=%d ok=%v", code, ok)
	}
	if _, ok := domainVerdictToStatusCode(domain.VerdictUnknown); ok {
		t.Fatal("UNKNOWN not writable")
	}
	if _, ok := domainVerdictToStatusCode(domain.VerdictRT); ok {
		t.Fatal("RT not writable")
	}
	if _, ok := domainVerdictToStatusCode(domain.VerdictRU); ok {
		t.Fatal("RU not writable")
	}
}
