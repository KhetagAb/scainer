package analyze

import (
	"fmt"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

func ScopeKeySubmission(u domain.StandaloneUnit) contests.ScopeKey {
	return contests.ScopeKey(fmt.Sprintf("submission:%s", u.Sub.ID))
}

func ScopeKeyProblem(u domain.ProblemUnit) contests.ScopeKey {
	return contests.ScopeKey(fmt.Sprintf("problem:%s:%s", u.Problem, u.Lang))
}

func ScopeKeyProblemParticipant(u domain.ProblemParticipantUnit) contests.ScopeKey {
	return contests.ScopeKey(fmt.Sprintf("problem-participant:%s:%s", u.Problem, u.Participant))
}
