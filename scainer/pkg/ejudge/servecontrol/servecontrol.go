package servecontrol

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"regexp"
	"strings"
	"time"

	ejauth "scainer/pkg/ejudge/auth"
)

var contestRowRe = regexp.MustCompile(`<td>(\d{4,6})</td>\s*<td>([^<]+)</td>`)

type Brief struct {
	ID   int
	Name string
}

// ListContests читает HTML-таблицу serve-control через browser session (auth.Session).
func ListContests(ctx context.Context, session ejauth.Session, timeout time.Duration) ([]Brief, error) {
	client := session.HTTPClient(timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, session.BaseURL()+"/cgi-bin/serve-control?SID="+session.SID, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	return ParseContestsHTML(body)
}

func ParseContestsHTML(body []byte) ([]Brief, error) {
	matches := contestRowRe.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("ejudge serve-control: contest table not found")
	}

	out := make([]Brief, 0, len(matches))
	seen := make(map[int]bool, len(matches))
	for _, m := range matches {
		id, err := strconv.Atoi(string(m[1]))
		if err != nil || id <= 0 {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Brief{
			ID:   id,
			Name: strings.TrimSpace(string(m[2])),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("ejudge serve-control: no contests parsed")
	}
	return out, nil
}
