package ejudge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	ejgen "scainer/generated/ejudge"
	"scainer/internal/domain"
)

type RunInfo struct {
	RunID   int
	Status  int
	Verdict domain.Verdict
}

type RunMessage struct {
	ClarID  int
	From    string // display name from view-source HTML
	Subject string
	Text    string
	Time    time.Time
}

func (c *Client) RunStatus(ctx context.Context, contestID, runID int) (RunInfo, error) {
	rid := runID
	body, err := c.masterJSON(ctx, &ejgen.MasterJSONParams{
		Json:      ejgen.MasterJSONParamsJsonN1,
		Action:    ejgen.RunStatusJson,
		ContestId: contestID,
		RunId:     &rid,
	})
	if err != nil {
		return RunInfo{}, err
	}
	reply, err := decodeReply[ejgen.RunStatusReply](body)
	if err != nil {
		return RunInfo{}, fmt.Errorf("ejudge run-status: json: %w", err)
	}
	if err := EnsureOK(reply.Ok, reply.Error); err != nil {
		return RunInfo{}, err
	}
	if reply.Result == nil || reply.Result.Run == nil {
		return RunInfo{}, fmt.Errorf("ejudge run-status: пустой result.run")
	}
	run := reply.Result.Run
	info := RunInfo{RunID: runID}
	if run.RunId != nil {
		info.RunID = *run.RunId
	}
	if run.Status != nil {
		info.Status = *run.Status
	}
	if run.StatusStr != nil && *run.StatusStr != "" {
		info.Verdict = domain.ParseVerdict(*run.StatusStr)
	} else {
		info.Verdict = VerdictFromStatus(info.Status)
	}
	return info, nil
}

// RunMessages читает комментарии со страницы view-source (action=36).
// Privileged run-messages-json в ejudge закомментирован / непригоден для судьи.
func (c *Client) RunMessages(ctx context.Context, contestID, runID int) ([]RunMessage, error) {
	body, err := c.viewSource(ctx, contestID, runID)
	if err != nil {
		return nil, err
	}
	msgs, err := parseViewSourceComments(body)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (c *Client) viewSource(ctx context.Context, contestID, runID int) ([]byte, error) {
	if c == nil || c.httpClient == nil {
		return nil, fmt.Errorf("ejudge: клиент не инициализирован")
	}
	u, err := url.Parse(c.baseURL + "/cgi-bin/new-master")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("action", actionViewSource)
	q.Set("contest_id", strconv.Itoa(contestID))
	q.Set("run_id", strconv.Itoa(runID))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ejudge view-source: HTTP %d: %s", resp.StatusCode, truncateBody(body, 200))
	}
	if len(body) > 0 && body[0] == '{' {
		return nil, fmt.Errorf("ejudge view-source: ожидался HTML, получен JSON: %s", truncateBody(body, 200))
	}
	return body, nil
}

func (c *Client) SendRunComment(ctx context.Context, contestID, runID int, text string) error {
	rid := runID
	msg := text
	body, err := c.masterForm(ctx, ejgen.MasterFormFormdataRequestBody{
		Json:      ejgen.MasterFormFormdataBodyJsonN1,
		Action:    actionSendRunComment,
		ContestId: contestID,
		RunId:     &rid,
		MsgText:   &msg,
	})
	if err != nil {
		return err
	}
	return ensureWriteOK(body, "send-run-comment")
}

func (c *Client) ChangeRunStatus(ctx context.Context, contestID, runID int, verdict domain.Verdict) error {
	code, ok := StatusCode(verdict)
	if !ok {
		return fmt.Errorf("ejudge: неизвестный вердикт %q", verdict)
	}
	rid := runID
	status := code
	body, err := c.masterForm(ctx, ejgen.MasterFormFormdataRequestBody{
		Json:      ejgen.MasterFormFormdataBodyJsonN1,
		Action:    actionChangeRunStatus,
		ContestId: contestID,
		RunId:     &rid,
		Status:    &status,
	})
	if err != nil {
		return err
	}
	return ensureWriteOK(body, "change-run-status")
}

func ensureWriteOK(body []byte, op string) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}
	var wrap ejgen.JSONEnvelope
	if err := json.Unmarshal(body, &wrap); err != nil {
		// HTML-ответ без JSON envelope (legacy CGI) — считаем успехом при HTTP 200.
		return nil
	}
	if wrap.Ok == nil && wrap.Error == nil {
		return nil
	}
	if err := EnsureOK(wrap.Ok, wrap.Error); err != nil {
		return fmt.Errorf("ejudge %s: %w", op, err)
	}
	return nil
}
