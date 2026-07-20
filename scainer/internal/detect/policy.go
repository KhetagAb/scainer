package detect

import "scainer/internal/domain"

type AnalysisPolicy struct {
	ExcludedProblems map[domain.ProblemID]bool
}

func (p AnalysisPolicy) Excluded(problem domain.ProblemID) bool {
	return p.ExcludedProblems[problem]
}
