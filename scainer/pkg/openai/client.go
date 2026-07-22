// Package openai — OpenAI-совместимый HTTP API (Ollama /v1/*).
// Auth: HTTP Basic (прокси) или Bearer.
package openai

import (
	"context"
	"fmt"
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
}

func New(baseURL, username, password string, timeout time.Duration) (*Client, error) {
	return newClient(baseURL, timeout, basicAuth(username, password))
}

func NewWithAPIKey(baseURL, apiKey string, timeout time.Duration) (*Client, error) {
	return newClient(baseURL, timeout, bearerAuth(apiKey))
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
	}, nil
}

func (c *Client) Timeout() time.Duration { return c.http.Timeout }

func (c *Client) BaseURL() string { return c.baseURL }

func basicAuth(user, pass string) oaigen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.SetBasicAuth(user, pass)
		return nil
	}
}

func bearerAuth(apiKey string) oaigen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+apiKey)
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
