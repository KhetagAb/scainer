package domain

type Verdict string

const (
	VerdictOK      Verdict = "OK"
	VerdictTL      Verdict = "TL"
	VerdictML      Verdict = "ML"
	VerdictWA      Verdict = "WA"
	VerdictPR      Verdict = "PR"
	VerdictCF      Verdict = "CF"
	VerdictCE      Verdict = "CE"
	VerdictAC      Verdict = "AC"
	VerdictDQ      Verdict = "DQ"
	VerdictRJ      Verdict = "RJ"
	VerdictUnknown Verdict = "UNKNOWN"
)

func ParseVerdict(s string) Verdict {
	switch Verdict(s) {
	case VerdictOK, VerdictTL, VerdictML, VerdictWA, VerdictPR, VerdictCF, VerdictCE, VerdictAC, VerdictDQ, VerdictRJ:
		return Verdict(s)
	default:
		return VerdictUnknown
	}
}
