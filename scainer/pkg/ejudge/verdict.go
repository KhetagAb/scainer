package ejudge

import "strings"

type Verdict string

const (
	VerdictOK      Verdict = "OK" // Accepted (tests passed)
	VerdictTL      Verdict = "TL" // Time Limit Exceeded
	VerdictML      Verdict = "ML" // Memory Limit Exceeded
	VerdictWA      Verdict = "WA" // Wrong Answer
	VerdictPR      Verdict = "PR" // Presentation Error
	VerdictCF      Verdict = "CF" // Check Failed
	VerdictCE      Verdict = "CE" // Compilation Error
	VerdictAC      Verdict = "AC" // Accepted (manual / off-line)
	VerdictDQ      Verdict = "DQ" // Disqualified
	VerdictRJ      Verdict = "RJ" // Rejected (incl. style violation)
	VerdictUnknown Verdict = "UNKNOWN"
)

func ParseVerdict(s string) Verdict {
	s = strings.TrimSpace(s)
	if alias := verdictAlias(s); alias != "" {
		s = alias
	}
	switch Verdict(s) {
	case VerdictOK, VerdictTL, VerdictML, VerdictWA, VerdictPR, VerdictCF, VerdictCE, VerdictAC, VerdictDQ, VerdictRJ:
		return Verdict(s)
	default:
		return VerdictUnknown
	}
}

func verdictAlias(s string) string {
	switch strings.ToUpper(s) {
	case "SV":
		return string(VerdictRJ)
	}
	return ""
}

// VerdictFromStatus — запасной маппинг по числовому status ejudge, если status_str пуст.
func VerdictFromStatus(code int) Verdict {
	switch code {
	case 0: // RUN_OK
		return VerdictOK
	case 1: // RUN_COMPILE_ERR
		return VerdictCE
	case 3: // RUN_TIME_LIMIT_ERR
		return VerdictTL
	case 5: // RUN_WRONG_ANSWER_ERR
		return VerdictWA
	case 6: // RUN_CHECK_FAILED
		return VerdictCF
	case 8: // RUN_ACCEPTED
		return VerdictAC
	case 10: // RUN_DISQUALIFIED
		return VerdictDQ
	case 12: // RUN_MEM_LIMIT_ERR
		return VerdictML
	case 14: // RUN_STYLE_ERR (SV)
		return VerdictRJ
	case 16: // RUN_PRESENTATION_ERR
		return VerdictPR
	case 17: // RUN_REJECTED
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
