package ejudge

import (
	"strings"

	"scainer/internal/domain"
)

func mapToDomainVerdict(status int, statusStr string) domain.Verdict {
	if statusStr != "" {
		return parseEjudgeStatusStr(statusStr)
	}
	return verdictFromEjudgeStatus(status)
}

func parseEjudgeStatusStr(s string) domain.Verdict {
	s = strings.TrimSpace(s)
	if alias := ejudgeStatusAlias(s); alias != "" {
		s = alias
	}
	switch domain.Verdict(s) {
	case domain.VerdictOK, domain.VerdictTL, domain.VerdictML, domain.VerdictWA, domain.VerdictPR,
		domain.VerdictCF, domain.VerdictCE, domain.VerdictDQ, domain.VerdictRJ,
		domain.VerdictRT, domain.VerdictIG, domain.VerdictPE, domain.VerdictRU:
		return domain.Verdict(s)
	default:
		return domain.VerdictUnknown
	}
}

func ejudgeStatusAlias(s string) string {
	switch strings.ToUpper(s) {
	case "SV":
		return string(domain.VerdictRJ)
	case "PT":
		return string(domain.VerdictWA)
	case "CG", "CD", "AV", "VS", "VT", "EM":
		return string(domain.VerdictRU)
	}
	return ""
}

func isRunningEjudgeStatus(code int) bool {
	switch code {
	case 20, 21, 22, 95, 96, 97, 98, 99:
		return true
	default:
		return false
	}
}

// verdictFromEjudgeStatus — запасной маппинг по числовому status ejudge, если status_str пуст.
func verdictFromEjudgeStatus(code int) domain.Verdict {
	if isRunningEjudgeStatus(code) {
		return domain.VerdictRU
	}
	switch code {
	case 0: // OK
		return domain.VerdictOK
	case 1: // CE
		return domain.VerdictCE
	case 2: // RT
		return domain.VerdictRT
	case 3: // TL
		return domain.VerdictTL
	case 4: // PE
		return domain.VerdictPE
	case 5: // WA
		return domain.VerdictWA
	case 6: // CF
		return domain.VerdictCF
	case 7: // PT — partial solution
		return domain.VerdictWA
	case 9: // IG
		return domain.VerdictIG
	case 10: // DQ
		return domain.VerdictDQ
	case 12: // ML
		return domain.VerdictML
	case 14: // SV
		return domain.VerdictRJ
	case 16: // PR
		return domain.VerdictPR
	case 17: // RJ
		return domain.VerdictRJ
	default:
		return domain.VerdictUnknown
	}
}

func domainVerdictToStatusCode(v domain.Verdict) (int, bool) {
	switch v {
	case domain.VerdictOK:
		return 0, true
	case domain.VerdictTL:
		return 3, true
	case domain.VerdictWA:
		return 5, true
	case domain.VerdictCF:
		return 6, true
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
