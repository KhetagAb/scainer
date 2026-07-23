package ejudge

import (
	"scainer/internal/domain"
	ejudgeapi "scainer/pkg/ejudge"
)

func toDomainVerdict(v ejudgeapi.Verdict) domain.Verdict {
	return domain.Verdict(v)
}

func toEjudgeVerdict(v domain.Verdict) (ejudgeapi.Verdict, bool) {
	ev := ejudgeapi.Verdict(v)
	if _, ok := ejudgeapi.StatusCode(ev); ok {
		return ev, true
	}
	return ev, false
}
