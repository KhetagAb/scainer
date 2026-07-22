// Package ejudge — HTTP-клиент ejudge (Bearer AQAA<token>, cgi-bin).
package ejudge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ejgen "scainer/generated/ejudge"
)

//go:generate go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

const (
	authPrefix     = "AQAA"
	defaultTimeout = 15 * time.Second

	actionSendRunComment  = "64"
	actionChangeRunStatus = "67"
	actionViewSource      = "36"
)

type Client struct {
	*ejgen.ClientWithResponses
	baseURL    string
	httpClient *http.Client
	authHeader string
	timeout    time.Duration
}

func New(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	httpClient := &http.Client{Timeout: timeout}
	base := strings.TrimRight(baseURL, "/")
	authHeader := "Bearer " + authPrefix + apiKey
	auth := func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", authHeader)
		return nil
	}

	api, err := ejgen.NewClientWithResponses(
		base,
		ejgen.WithHTTPClient(httpClient),
		ejgen.WithRequestEditorFn(auth),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		ClientWithResponses: api,
		baseURL:             base,
		httpClient:          httpClient,
		authHeader:          authHeader,
		timeout:             timeout,
	}, nil
}

func (c *Client) Timeout() time.Duration { return c.timeout }

func EnsureOK(ok *bool, apiErr *ejgen.Error) error {
	if ok != nil && *ok {
		return nil
	}
	if apiErr != nil {
		switch {
		case apiErr.Message != nil && *apiErr.Message != "":
			return fmt.Errorf("ejudge: %s", *apiErr.Message)
		case apiErr.Symbol != nil && *apiErr.Symbol != "":
			return fmt.Errorf("ejudge: %s", *apiErr.Symbol)
		}
	}
	return fmt.Errorf("ejudge: ok=false")
}

func DownloadRunParams(contestID, runID int) *ejgen.DownloadRunParams {
	noDisp := 1
	return &ejgen.DownloadRunParams{
		Json:      ejgen.DownloadRunParamsJsonN1,
		Action:    ejgen.DownloadRun,
		ContestId: contestID,
		RunId:     runID,
		NoDisp:    &noDisp,
	}
}

func (c *Client) masterJSON(ctx context.Context, params *ejgen.MasterJSONParams) ([]byte, error) {
	if c == nil || c.ClientWithResponses == nil {
		return nil, fmt.Errorf("ejudge: клиент не инициализирован")
	}
	resp, err := c.MasterJSONWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("ejudge %s: HTTP %d: %s", params.Action, resp.StatusCode(), truncateBody(resp.Body, 200))
	}
	return resp.Body, nil
}

func (c *Client) masterForm(ctx context.Context, body ejgen.MasterFormFormdataRequestBody) ([]byte, error) {
	if c == nil || c.ClientWithResponses == nil {
		return nil, fmt.Errorf("ejudge: клиент не инициализирован")
	}
	// WithResponse парсит тело как JSON и падает на пустом ответе ejudge (action 64/67).
	resp, err := c.MasterFormWithFormdataBody(ctx, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ejudge action=%s: HTTP %d: %s", body.Action, resp.StatusCode, truncateBody(raw, 200))
	}
	return raw, nil
}

func decodeReply[T any](body []byte) (T, error) {
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}
	return out, nil
}

func truncateBody(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
