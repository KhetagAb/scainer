// Package ejudge — HTTP API ejudge. Auth: Bearer AQAA<token>.
package ejudge

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	ejgen "github.com/lksh/scainer/generated/ejudge"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 -config api/oapi-codegen.yaml api/openapi.yaml

const (
	authPrefix     = "AQAA"
	defaultTimeout = 15 * time.Second
)

type Client struct {
	*ejgen.ClientWithResponses
	http    *http.Client
	baseURL string
	editors []ejgen.RequestEditorFn
}

func New(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	httpClient := &http.Client{Timeout: timeout}
	base := strings.TrimRight(baseURL, "/")
	auth := func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+authPrefix+apiKey)
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
		http:                httpClient,
		baseURL:             base,
		editors:             []ejgen.RequestEditorFn{auth},
	}, nil
}

func (c *Client) Timeout() time.Duration { return c.http.Timeout }

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

func ListRunsParams(contestID int, firstRun, lastRun *int) *ejgen.ListRunsParams {
	return &ejgen.ListRunsParams{
		Json:      ejgen.ListRunsParamsJsonN1,
		Action:    ejgen.ListRunsJson,
		ContestId: contestID,
		FirstRun:  firstRun,
		LastRun:   lastRun,
	}
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
