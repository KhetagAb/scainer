package ejudge

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"
)

var ejudgeWallClock = time.FixedZone("MSK", 3*60*60)

var (
	reRunCommentsSection = regexp.MustCompile(`(?is)<h2>\s*Run comments\s*</h2>(.*?)(?:<h2>\s*Run comments for previous runs\s*</h2>|<h2>\s*Add a new run comment\s*</h2>|$)`)
	reMessageRow         = regexp.MustCompile(`(?is)<tr>\s*<td\s+class="profile">\s*<b>(.*?)</b>\s*<br\s*/?>\s*([0-9]{4}-[0-9]{2}-[0-9]{2}\s+[0-9]{2}:[0-9]{2}:[0-9]{2}).*?</td>\s*<td>\s*<pre>(.*?)</pre>`)
)

func parseViewSourceComments(body []byte) ([]RunMessage, error) {
	section := reRunCommentsSection.FindSubmatch(body)
	if section == nil {
		return nil, nil
	}
	rows := reMessageRow.FindAllSubmatch(section[1], -1)
	out := make([]RunMessage, 0, len(rows))
	for i, row := range rows {
		author := strings.TrimSpace(html.UnescapeString(string(row[1])))
		ts, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(string(row[2])), ejudgeWallClock)
		if err != nil {
			return nil, fmt.Errorf("ejudge view-source: time %q: %w", row[2], err)
		}
		text := strings.TrimSuffix(html.UnescapeString(string(row[3])), "\n")
		out = append(out, RunMessage{
			ClarID: i + 1,
			From:   author,
			Text:   text,
			Time:   ts,
		})
	}
	return out, nil
}
