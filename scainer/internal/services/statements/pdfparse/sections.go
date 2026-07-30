package pdfparse

import (
	"regexp"
	"strings"
)

var problemHeaderRE = regexp.MustCompile(`(?i)Задача\s*([A-Za-zА-Яа-яЁё0-9]+)\s*\.`)

type ProblemSection struct {
	Label     string
	StartPage int // 1-based
}

func normalizeProblemLabel(label string) string {
	return strings.ToUpper(strings.TrimSpace(label))
}

func findProblemSections(pageTexts []string) []ProblemSection {
	var sections []ProblemSection
	for i, text := range pageTexts {
		page := i + 1
		seen := make(map[string]struct{})
		loc := problemHeaderRE.FindAllStringSubmatchIndex(text, -1)
		for _, match := range loc {
			if len(match) < 4 {
				continue
			}
			label := normalizeProblemLabel(text[match[2]:match[3]])
			if label == "" {
				continue
			}
			if _, ok := seen[label]; ok {
				continue
			}
			seen[label] = struct{}{}
			sections = append(sections, ProblemSection{Label: label, StartPage: page})
		}
	}
	return sections
}

type PageRange struct {
	StartPage int
	EndPage   int
}

func pagesForProblem(numPages int, sections []ProblemSection, problemLabel string) (PageRange, bool) {
	trimmed := strings.TrimSpace(problemLabel)
	if trimmed == "" {
		return PageRange{}, false
	}
	key := normalizeProblemLabel(trimmed)
	idx := -1
	for i, s := range sections {
		if s.Label == key {
			idx = i
			break
		}
	}
	if idx < 0 {
		if dash := strings.Index(trimmed, " - "); dash > 0 {
			short := normalizeProblemLabel(trimmed[:dash])
			for i, s := range sections {
				if s.Label == short {
					idx = i
					break
				}
			}
		}
	}
	if idx < 0 {
		return PageRange{}, false
	}
	start := sections[idx].StartPage
	end := numPages
	if idx+1 < len(sections) {
		end = sections[idx+1].StartPage - 1
	}
	if start > end || start < 1 || end > numPages {
		return PageRange{}, false
	}
	return PageRange{StartPage: start, EndPage: end}, true
}
