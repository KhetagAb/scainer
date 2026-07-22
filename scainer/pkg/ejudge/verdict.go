package ejudge

import "scainer/internal/domain"

// VerdictFromStatus — запасной маппинг по числовому status ejudge, если status_str пуст.
func VerdictFromStatus(code int) domain.Verdict {
	switch code {
	case 0:
		return domain.VerdictOK
	case 3:
		return domain.VerdictTL
	case 5:
		return domain.VerdictWA
	case 6:
		return domain.VerdictCF
	case 8:
		return domain.VerdictAC
	case 10:
		return domain.VerdictDQ
	case 12:
		return domain.VerdictML
	case 16:
		return domain.VerdictPR
	case 17:
		return domain.VerdictRJ
	default:
		return domain.VerdictUnknown
	}
}

func StatusCode(v domain.Verdict) (int, bool) {
	switch v {
	case domain.VerdictOK:
		return 0, true
	case domain.VerdictTL:
		return 3, true
	case domain.VerdictWA:
		return 5, true
	case domain.VerdictCF:
		return 6, true
	case domain.VerdictAC:
		return 8, true
	case domain.VerdictDQ:
		return 10, true
	case domain.VerdictML:
		return 12, true
	case domain.VerdictPR:
		return 16, true
	case domain.VerdictRJ:
		return 17, true
	default:
		return 0, false
	}
}
