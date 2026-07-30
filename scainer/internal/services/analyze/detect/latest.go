package detect

import "scainer/internal/domain"

// LatestPerParticipant — по одной самой новой посылке на участника.
// Порядок: первое появление участника во входе. При равном времени побеждает больший ID.
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

// LatestOkPerParticipant — LatestPerParticipant по OK и PR (pending review).
func LatestOkPerParticipant(subs []domain.Submission) []domain.Submission {
	return LatestPerParticipant(filterOK(subs))
}

// filterOK оставляет посылки с VerdictOK или VerdictPR, сохраняя порядок входа.
func filterOK(subs []domain.Submission) []domain.Submission {
	ok := make([]domain.Submission, 0, len(subs))
	for _, s := range subs {
		if s.Verdict == domain.VerdictOK || s.Verdict == domain.VerdictPR {
			ok = append(ok, s)
		}
	}
	return ok
}

// isNewerSubmission: a новее b по SubmittedAt; при равенстве — по большему ID.
func isNewerSubmission(a, b domain.Submission) bool {
	if a.SubmittedAt.After(b.SubmittedAt) {
		return true
	}
	if a.SubmittedAt.Before(b.SubmittedAt) {
		return false
	}
	return a.ID > b.ID
}
