package ejudge

type Verdict string

const (
	VerdictOK      Verdict = "OK"
	VerdictTL      Verdict = "TL"
	VerdictML      Verdict = "ML"
	VerdictWA      Verdict = "WA"
	VerdictPR      Verdict = "PR"
	VerdictCF      Verdict = "CF"
	VerdictAC      Verdict = "AC"
	VerdictDQ      Verdict = "DQ"
	VerdictRJ      Verdict = "RJ"
	VerdictUnknown Verdict = "UNKNOWN"
)

func ParseVerdict(s string) Verdict {
	switch Verdict(s) {
	case VerdictOK, VerdictTL, VerdictML, VerdictWA, VerdictPR, VerdictCF, VerdictAC, VerdictDQ, VerdictRJ:
		return Verdict(s)
	default:
		return VerdictUnknown
	}
}

// VerdictFromStatus — запасной маппинг по числовому status ejudge, если status_str пуст.
func VerdictFromStatus(code int) Verdict {
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
	case 12:
		return VerdictML
	case 16:
		return VerdictPR
	case 17:
		return VerdictRJ
	default:
		return VerdictUnknown
	}
}

func StatusCode(v Verdict) (int, bool) {
	switch v {
	case VerdictOK:
		return 0, true
	case VerdictTL:
		return 3, true
	case VerdictWA:
		return 5, true
	case VerdictCF:
		return 6, true
	case VerdictAC:
		return 8, true
	case VerdictDQ:
		return 10, true
	case VerdictML:
		return 12, true
	case VerdictPR:
		return 16, true
	case VerdictRJ:
		return 17, true
	default:
		return 0, false
	}
}
