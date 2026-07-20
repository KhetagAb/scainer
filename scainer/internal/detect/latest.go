package detect

import "github.com/lksh/scainer/internal/domain"

func LatestPerParticipant(subs []domain.Submission) []domain.Submission {
	best := make(map[domain.ParticipantID]domain.Submission, len(subs))
	order := make([]domain.ParticipantID, 0)
	for _, s := range subs {
		prev, ok := best[s.Participant]
		if !ok {
			best[s.Participant] = s
			order = append(order, s.Participant)
			continue
		}
		if isNewerSubmission(s, prev) {
			best[s.Participant] = s
		}
	}
	out := make([]domain.Submission, 0, len(order))
	for _, p := range order {
		out = append(out, best[p])
	}
	return out
}

func LatestSuccessfulPerParticipant(subs []domain.Submission) []domain.Submission {
	ok := make([]domain.Submission, 0, len(subs))
	for _, s := range subs {
		if s.Verdict == domain.VerdictOK {
			ok = append(ok, s)
		}
	}
	return LatestPerParticipant(ok)
}

func isNewerSubmission(a, b domain.Submission) bool {
	if a.SubmittedAt.After(b.SubmittedAt) {
		return true
	}
	if a.SubmittedAt.Before(b.SubmittedAt) {
		return false
	}
	return a.ID > b.ID
}
