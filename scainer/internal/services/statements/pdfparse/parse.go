package pdfparse

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"

	"scainer/internal/domain"
)

type ParseResult struct {
	Statements map[domain.ProblemID]domain.ProblemStatement
	PageRanges map[domain.ProblemID]PageRange
}

func ParseContestPDF(r io.Reader) (ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return ParseResult{}, fmt.Errorf("pdfparse: read: %w", err)
	}
	pageTexts, err := extractPageTexts(data)
	if err != nil {
		return ParseResult{}, err
	}
	if len(pageTexts) == 0 {
		return ParseResult{}, fmt.Errorf("pdfparse: пустой PDF")
	}

	sections := findProblemSections(pageTexts)
	if len(sections) == 0 {
		return ParseResult{}, fmt.Errorf("pdfparse: задачи не найдены")
	}

	out := ParseResult{
		Statements: make(map[domain.ProblemID]domain.ProblemStatement, len(sections)),
		PageRanges: make(map[domain.ProblemID]PageRange, len(sections)),
	}

	for i, sec := range sections {
		endPage := len(pageTexts)
		if i+1 < len(sections) {
			endPage = sections[i+1].StartPage - 1
		}
		var chunk strings.Builder
		for p := sec.StartPage; p <= endPage && p <= len(pageTexts); p++ {
			if chunk.Len() > 0 {
				chunk.WriteByte(' ')
			}
			chunk.WriteString(pageTexts[p-1])
		}
		pid := domain.ProblemID(sec.Label)
		ps := parseProblemFields(pid, chunk.String())
		ps.Problem = pid
		out.Statements[pid] = ps
		out.PageRanges[pid] = PageRange{StartPage: sec.StartPage, EndPage: endPage}
	}
	return out, nil
}

func extractPageTexts(data []byte) ([]string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("pdfparse: open: %w", err)
	}
	var pages []string
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			pages = append(pages, "")
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, fmt.Errorf("pdfparse: page %d: %w", i, err)
		}
		pages = append(pages, collapseWS(text))
	}
	return pages, nil
}
