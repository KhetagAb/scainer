package contests

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"scainer/internal/domain"
	"scainer/pkg/ejudge/servecontrol"
)

const (
	skipReasonTemplate        = "шаблон"
	skipReasonParallelUnknown = "параллель не распознана"
)

var parallelPattern = regexp.MustCompile(`(?i)Параллель\s+(10|[3-9]|X|F|R)\b`)
var templatePattern = regexp.MustCompile(`(?i)template`)

type CatalogEntry struct {
	ID         string
	Name       string
	ParallelID string
	Reason     string
}

type EjudgeImportResult struct {
	Added      []CatalogEntry
	Duplicates []CatalogEntry
	Skipped    []CatalogEntry
}

type EjudgeContestLister func(ctx context.Context) ([]servecontrol.Brief, error)

func ClassifyEjudgeContests(briefs []servecontrol.Brief) (toImport, skipped []CatalogEntry) {
	for _, brief := range briefs {
		id := strconv.Itoa(brief.ID)
		name := strings.TrimSpace(brief.Name)
		if name == "" {
			continue
		}
		if templatePattern.MatchString(name) {
			skipped = append(skipped, CatalogEntry{ID: id, Name: name, Reason: skipReasonTemplate})
			continue
		}
		match := parallelPattern.FindStringSubmatch(name)
		if match == nil {
			skipped = append(skipped, CatalogEntry{ID: id, Name: name, Reason: skipReasonParallelUnknown})
			continue
		}
		parallelID := match[1]
		if parallelID != "" && parallelID[0] >= '0' && parallelID[0] <= '9' {
			// numeric parallel id as-is
		} else {
			parallelID = strings.ToUpper(parallelID)
		}
		toImport = append(toImport, CatalogEntry{
			ID:         id,
			Name:       name,
			ParallelID: parallelID,
		})
	}
	return toImport, skipped
}

func (s *Service) ImportFromEjudge(ctx context.Context, list EjudgeContestLister) (EjudgeImportResult, error) {
	if list == nil {
		return EjudgeImportResult{}, errors.New("ejudge contest lister is not configured")
	}
	briefs, err := list(ctx)
	if err != nil {
		return EjudgeImportResult{}, err
	}

	toImport, skipped := ClassifyEjudgeContests(briefs)
	result := EjudgeImportResult{Skipped: skipped}

	for _, entry := range toImport {
		_, err := s.Register(ctx, Registration{
			ID:         domain.ContestID(entry.ID),
			ParallelID: entry.ParallelID,
		})
		if errors.Is(err, ErrDuplicateContest) {
			result.Duplicates = append(result.Duplicates, entry)
			continue
		}
		if err != nil {
			return EjudgeImportResult{}, err
		}
		result.Added = append(result.Added, entry)
	}
	return result, nil
}
