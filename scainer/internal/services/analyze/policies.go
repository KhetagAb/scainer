package analyze

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

type Once[U domain.Unit] struct {
	Key func(U) contests.ScopeKey
}

func (p Once[U]) ScopeKey(u U) contests.ScopeKey {
	return p.Key(u)
}

func (p Once[U]) ShouldRun(_ context.Context, u U, prev contests.DetectorProgress) bool {
	_, ok := prev[p.ScopeKey(u)]
	return !ok
}

func (p Once[U]) AfterRun(u U, next *contests.DetectorProgress) {
	(*next)[p.ScopeKey(u)] = nil
}

type UnlessChanged[U domain.Unit] struct {
	Key    func(U) contests.ScopeKey
	SubIDs func(U) []domain.SubmissionID
}

func (p UnlessChanged[U]) ScopeKey(u U) contests.ScopeKey {
	return p.Key(u)
}

func (p UnlessChanged[U]) ShouldRun(_ context.Context, u U, prev contests.DetectorProgress) bool {
	baseline, ok := prev[p.ScopeKey(u)]
	if !ok {
		return true
	}
	return !sameSubmissionSet(baseline, p.SubIDs(u))
}

func (p UnlessChanged[U]) AfterRun(u U, next *contests.DetectorProgress) {
	ids := p.SubIDs(u)
	(*next)[p.ScopeKey(u)] = append([]domain.SubmissionID(nil), ids...)
}

func SubmissionIDsFromProblemUnit(u domain.ProblemUnit) []domain.SubmissionID {
	ids := make([]domain.SubmissionID, len(u.Subs))
	for i, sub := range u.Subs {
		ids[i] = sub.ID
	}
	return ids
}

type ManualOnly[U domain.Unit] struct {
	Key func(U) contests.ScopeKey
}

func (p ManualOnly[U]) ScopeKey(u U) contests.ScopeKey {
	return p.Key(u)
}

func (p ManualOnly[U]) ShouldRun(ctx context.Context, _ U, _ contests.DetectorProgress) bool {
	return ManualFromContext(ctx)
}

func (ManualOnly[U]) AfterRun(U, *contests.DetectorProgress) {}

func sameSubmissionSet(a, b []domain.SubmissionID) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[domain.SubmissionID]int, len(a))
	for _, id := range a {
		counts[id]++
	}
	for _, id := range b {
		if counts[id] == 0 {
			return false
		}
		counts[id]--
	}
	return true
}
