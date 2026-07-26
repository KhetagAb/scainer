package contests

import (
	"slices"

	"scainer/internal/domain"
	"scainer/internal/services/scoring"
)

func FlattenSignals(signals Signals) []domain.Signal {
	if len(signals) == 0 {
		return nil
	}

	detNames := make([]DetectorName, 0, len(signals))
	for name := range signals {
		detNames = append(detNames, name)
	}
	slices.Sort(detNames)

	var out []domain.Signal
	for _, name := range detNames {
		byKey := signals[name]
		keys := make([]ScopeKey, 0, len(byKey))
		for key := range byKey {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			out = append(out, byKey[key]...)
		}
	}
	return out
}

func (snap *AnalysisSnapshot) Findings(scorer scoring.Scorer) []domain.Finding {
	if snap == nil || len(snap.Signals) == 0 {
		return nil
	}
	return scorer.Score(FlattenSignals(snap.Signals))
}

func (snap *AnalysisSnapshot) Clone() AnalysisSnapshot {
	if snap == nil {
		return AnalysisSnapshot{
			Progress: make(Progress),
			Signals:  make(Signals),
		}
	}
	next := AnalysisSnapshot{
		ContestID:  snap.ContestID,
		ComputedAt: snap.ComputedAt,
		Progress:   make(Progress, len(snap.Progress)),
		Signals:    make(Signals, len(snap.Signals)),
	}
	for det, byKey := range snap.Progress {
		cp := make(DetectorProgress, len(byKey))
		for key, ids := range byKey {
			if ids == nil {
				cp[key] = nil
				continue
			}
			slice := make(ScopeProgress, len(ids))
			copy(slice, ids)
			cp[key] = slice
		}
		next.Progress[det] = cp
	}
	for det, byKey := range snap.Signals {
		cp := make(DetectorSignals, len(byKey))
		for key, sigs := range byKey {
			if sigs != nil {
				slice := make([]domain.Signal, len(sigs))
				copy(slice, sigs)
				cp[key] = slice
			}
		}
		next.Signals[det] = cp
	}
	return next
}

func (snap *AnalysisSnapshot) DetectorMaps(name DetectorName) (DetectorProgress, DetectorSignals) {
	if snap.Progress[name] == nil {
		snap.Progress[name] = make(DetectorProgress)
	}
	if snap.Signals[name] == nil {
		snap.Signals[name] = make(DetectorSignals)
	}
	return snap.Progress[name], snap.Signals[name]
}
