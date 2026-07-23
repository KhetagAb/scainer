package ejudge

import (
	"bytes"
	stdhtml "html"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

var ejudgeWallClock = time.FixedZone("MSK", 3*60*60)

var (
	reProfileTimestamp = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	commentSectionTitles = []string{
		"Run comments",
		"Run comments for previous runs",
	}
)

// Семантический XPath: h2 заголовок секции → ближайшая table.message-table → строки с td.profile.
func commentRowsXPath(sectionTitle string) string {
	return fmt.Sprintf(
		`//h2[normalize-space()='%s']/following::div[contains(concat(' ', normalize-space(@class), ' '), ' width-100 ')][1]//tr[td[@class='profile']]`,
		sectionTitle,
	)
}

func parseViewSourceComments(body []byte) ([]RunMessage, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ejudge view-source: html: %w", err)
	}

	var out []RunMessage
	clarID := 1
	for _, title := range commentSectionTitles {
		rows, err := parseCommentSection(doc, commentRowsXPath(title), clarID)
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
		clarID += len(rows)
	}
	return out, nil
}

func parseCommentSection(doc *html.Node, xpath string, startClarID int) ([]RunMessage, error) {
	nodes, err := htmlquery.QueryAll(doc, xpath)
	if err != nil {
		return nil, fmt.Errorf("ejudge view-source: xpath %q: %w", xpath, err)
	}

	out := make([]RunMessage, 0, len(nodes))
	for i, tr := range nodes {
		msg, err := parseCommentRow(tr, startClarID+i)
		if err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, nil
}

func parseCommentRow(tr *html.Node, clarID int) (RunMessage, error) {
	profile := htmlquery.FindOne(tr, `./td[@class='profile']`)
	pre := htmlquery.FindOne(tr, `./td/pre`)
	if profile == nil || pre == nil {
		return RunMessage{}, fmt.Errorf("ejudge view-source: неполная строка комментария")
	}

	authorNode := htmlquery.FindOne(profile, `./b`)
	if authorNode == nil {
		return RunMessage{}, fmt.Errorf("ejudge view-source: автор не найден")
	}
	author := strings.TrimSpace(htmlquery.InnerText(authorNode))

	tsText := reProfileTimestamp.FindString(htmlquery.InnerText(profile))
	if tsText == "" {
		return RunMessage{}, fmt.Errorf("ejudge view-source: время не найдено у %q", author)
	}
	ts, err := time.ParseInLocation("2006-01-02 15:04:05", tsText, ejudgeWallClock)
	if err != nil {
		return RunMessage{}, fmt.Errorf("ejudge view-source: time %q: %w", tsText, err)
	}

	text := strings.TrimSuffix(stdhtml.UnescapeString(htmlquery.InnerText(pre)), "\n")
	return RunMessage{
		ClarID: clarID,
		From:   author,
		Text:   text,
		Time:   ts,
	}, nil
}
