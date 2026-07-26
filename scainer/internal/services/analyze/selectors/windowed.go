package selectors

import (
	"sort"

	"scainer/internal/domain"
)

// WindowEndingAtLastOK — до size посылок, последняя — last OK (по времени).
// Если до OK меньше size — возвращает все от начала до OK.
// Нет OK → nil. Посылки после last OK отбрасываются.
func WindowEndingAtLastOK(subs []domain.Submission, size int) []domain.Submission {
	if size <= 0 {
		return nil
	}
	sorted := sortBySubmittedAt(subs)
	okIdx := -1
	for i := len(sorted) - 1; i >= 0; i-- {
		if sorted[i].Verdict == domain.VerdictOK {
			okIdx = i
			break
		}
	}
	if okIdx < 0 {
		return nil
	}
	start := okIdx + 1 - size
	if start < 0 {
		start = 0
	}
	return append([]domain.Submission(nil), sorted[start:okIdx+1]...)
}

// sortBySubmittedAt — копия, по SubmittedAt asc; при равенстве — меньший ID раньше.
func sortBySubmittedAt(subs []domain.Submission) []domain.Submission {
	out := append([]domain.Submission(nil), subs...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SubmittedAt.Equal(out[j].SubmittedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].SubmittedAt.Before(out[j].SubmittedAt)
	})
	return out
}
