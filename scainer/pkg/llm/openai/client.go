// Package openai — OpenAI-совместимый HTTP API (Ollama /v1/*).
// Auth: HTTP Basic или заголовок Authorization: Api-Key.
package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	oaigen "scainer/generated/openai"
)

//go:generate go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

const defaultTimeout = 60 * time.Second

type Client struct {
	*oaigen.ClientWithResponses
	http    *http.Client
	baseURL string
	auth    oaigen.RequestEditorFn
}

func New(baseURL, username, password string, timeout time.Duration) (*Client, error) {
	return newClient(baseURL, timeout, basicAuth(username, password))
}

func NewWithAPIKey(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	return newClient(baseURL, timeout, apiKeyAuth(apiKey))
}

func newClient(baseURL string, timeout time.Duration, auth oaigen.RequestEditorFn) (*Client, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("openai: пустой base URL")
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	httpClient := &http.Client{Timeout: timeout}
	api, err := oaigen.NewClientWithResponses(
		base,
		oaigen.WithHTTPClient(httpClient),
		oaigen.WithRequestEditorFn(auth),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		ClientWithResponses: api,
		http:                httpClient,
		baseURL:             base,
		auth:                auth,
	}, nil
}

func (c *Client) HTTPClient() *http.Client { return c.http }

func (c *Client) BaseURL() string { return c.baseURL }

func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if c.auth != nil {
		if err := c.auth(ctx, req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

func (c *Client) Timeout() time.Duration { return c.http.Timeout }

func basicAuth(user, pass string) oaigen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.SetBasicAuth(user, pass)
		return nil
	}
}

func apiKeyAuth(apiKey string) oaigen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Api-Key "+apiKey)
		return nil
	}
}

// EnsureOK проверяет HTTP-статус ответа сгенерированного клиента.
func EnsureOK(status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Errorf("openai: HTTP %d", status)
	}
	if len(msg) > 200 {
		msg = msg[:200] + "…"
	}
	return fmt.Errorf("openai: HTTP %d: %s", status, msg)
}
