package ejudge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	ejgen "github.com/lksh/scainer/generated/ejudge"
)

type ContestInfo struct {
	ID   int
	Name string
}

// В openapi не заведён: path /cgi-bin/master уже занят download-run.
func (c *Client) ContestStatus(ctx context.Context, contestID int) (ContestInfo, error) {
	if c == nil || c.http == nil {
		return ContestInfo{}, fmt.Errorf("ejudge: клиент не инициализирован")
	}
	q := url.Values{
		"json":       {"1"},
		"action":     {"contest-status-json"},
		"contest_id": {strconv.Itoa(contestID)},
	}
	u := c.baseURL + "/cgi-bin/master?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ContestInfo{}, err
	}
	for _, ed := range c.editors {
		if err := ed(ctx, req); err != nil {
			return ContestInfo{}, err
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return ContestInfo{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ContestInfo{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return ContestInfo{}, fmt.Errorf("ejudge contest-status: HTTP %d: %s", resp.StatusCode, truncateBody(body, 200))
	}

	var wrap struct {
		Ok     *bool        `json:"ok"`
		Error  *ejgen.Error `json:"error"`
		Result *struct {
			Contest *struct {
				ID   *int    `json:"id"`
				Name *string `json:"name"`
			} `json:"contest"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return ContestInfo{}, fmt.Errorf("ejudge contest-status: json: %w", err)
	}
	if err := EnsureOK(wrap.Ok, wrap.Error); err != nil {
		return ContestInfo{}, err
	}
	if wrap.Result == nil || wrap.Result.Contest == nil {
		return ContestInfo{}, fmt.Errorf("ejudge contest-status: пустой result.contest")
	}
	info := ContestInfo{ID: contestID}
	if wrap.Result.Contest.ID != nil {
		info.ID = *wrap.Result.Contest.ID
	}
	if wrap.Result.Contest.Name != nil {
		info.Name = *wrap.Result.Contest.Name
	}
	return info, nil
}

func truncateBody(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
