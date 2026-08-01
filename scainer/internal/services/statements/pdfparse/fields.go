package pdfparse

import (
	"regexp"
	"strings"

	"scainer/internal/domain"
)

var titleRE = regexp.MustCompile(`(?i)Задача\s*([A-Za-zА-Яа-яЁё0-9]+)\s*\.\s*(.+?)\s*Имявходногофайла`)

const (
	markerExamples = "Примеры"
	markerNotes    = "Замечание"
)

func parseProblemFields(problemID domain.ProblemID, text string) domain.ProblemStatement {
	text = collapseWS(text)
	ps := domain.ProblemStatement{Problem: problemID}
	if m := titleRE.FindStringSubmatch(text); len(m) >= 3 {
		ps.Title = strings.TrimSpace(m[2])
	}

	exIdx := strings.Index(text, markerExamples)
	if exIdx < 0 {
		start := statementBodyStart(text)
		ps.RawStatement = strings.TrimSpace(text[start:])
		trimNextProblem(&ps.RawStatement)
		return ps
	}

	start := statementBodyStart(text[:exIdx])
	stmt := strings.TrimSpace(text[start:exIdx])
	rest := strings.TrimSpace(text[exIdx+len(markerExamples):])

	if noteIdx := strings.Index(rest, markerNotes); noteIdx >= 0 {
		ps.Examples = parseExamples(rest[:noteIdx])
		note := strings.TrimSpace(rest[noteIdx+len(markerNotes):])
		trimNextProblem(&note)
		if note != "" {
			stmt = strings.TrimSpace(stmt + "\n\n" + markerNotes + "\n" + note)
		}
	} else if k := problemHeaderRE.FindStringIndex(rest); k != nil {
		ps.Examples = parseExamples(rest[:k[0]])
	} else {
		ps.Examples = parseExamples(rest)
	}

	ps.RawStatement = stmt
	trimNextProblem(&ps.RawStatement)
	return ps
}

func statementBodyStart(prefix string) int {
	for _, marker := range []string{"мегабайт", "секунды", "секунд"} {
		if i := strings.Index(prefix, marker); i >= 0 {
			return i + len(marker)
		}
	}
	return 0
}

func trimNextProblem(s *string) {
	if k := problemHeaderRE.FindStringIndex(*s); k != nil && k[0] > 0 {
		*s = strings.TrimSpace((*s)[:k[0]])
	}
}

func parseExamples(block string) []domain.ProblemExample {
	block = strings.TrimSpace(block)
	if block == "" {
		return nil
	}
	for _, glued := range []string{"стандартныйввод", "стандартныйвывод", "стандартный ввод", "стандартный вывод"} {
		block = strings.ReplaceAll(block, glued, "")
	}
	parts := regexp.MustCompile(`\s{2,}`).Split(collapseWS(block), -1)
	var nums []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			nums = append(nums, p)
		}
	}
	if len(nums) < 2 {
		return []domain.ProblemExample{{Input: block, Output: ""}}
	}
	mid := (len(nums) + 1) / 2
	return []domain.ProblemExample{{
		Input:  strings.Join(nums[:mid], " "),
		Output: strings.Join(nums[mid:], " "),
	}}
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
