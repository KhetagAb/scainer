package domain

type Verdict string

const (
	VerdictOK      Verdict = "OK"
	VerdictTL      Verdict = "TL"
	VerdictML      Verdict = "ML"
	VerdictWA      Verdict = "WA"
	VerdictPD      Verdict = "PD"
	VerdictCF      Verdict = "CF"
	VerdictAC      Verdict = "AC"
	VerdictDQ      Verdict = "DQ"
	VerdictUnknown Verdict = "UNKNOWN"
)

func ParseVerdict(s string) Verdict {
	switch Verdict(s) {
	case VerdictOK, VerdictTL, VerdictML, VerdictWA, VerdictPD, VerdictCF, VerdictAC, VerdictDQ:
		return Verdict(s)
	default:
		return VerdictUnknown
	}
}

// Запасной маппинг по числовому status ejudge, если status_str пуст.
func VerdictFromEjudgeStatus(code int) Verdict {
	switch code {
	case 0:
		return VerdictOK
	case 3:
		return VerdictTL
	case 5:
		return VerdictWA
	case 6:
		return VerdictCF
	case 8:
		return VerdictAC
	case 10:
		return VerdictDQ
	case 11:
		return VerdictPD
	case 12:
		return VerdictML
	default:
		return VerdictUnknown
	}
}
