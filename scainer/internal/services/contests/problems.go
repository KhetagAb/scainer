package contests

import (
	"cmp"
	"slices"

	"scainer/internal/domain"
)

func buildProblemInfos(byProblem map[domain.ProblemID][]domain.Submission) []ProblemInfo {
	ids := make([]domain.ProblemID, 0, len(byProblem))
	for pid := range byProblem {
		ids = append(ids, pid)
	}
	slices.SortFunc(ids, func(a, b domain.ProblemID) int {
		return cmp.Compare(string(a), string(b))
	})

	out := make([]ProblemInfo, 0, len(ids))
	for _, pid := range ids {
		submissions := byProblem[pid]
		pending := 0
		for _, sub := range submissions {
			if sub.Verdict == domain.VerdictPR {
				pending++
			}
		}
		out = append(out, ProblemInfo{
			ID:              pid,
			Name:            problemDisplayName(pid, submissions),
			SubmissionCount: len(submissions),
			PendingCount:    pending,
		})
	}
	return out
}

func problemDisplayName(id domain.ProblemID, subs []domain.Submission) string {
	for _, sub := range subs {
		if sub.Meta == nil {
			continue
		}
		if name, ok := sub.Meta["problem_name"].(string); ok && name != "" {
			return name
		}
	}
	return string(id)
}
