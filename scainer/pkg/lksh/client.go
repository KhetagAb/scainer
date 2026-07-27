// Package lksh — открытый портал ejudge.lksh.ru (расписание параллели, PDF условий).
package lksh

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	portalURL  string
	httpClient *http.Client
}

func NewClient(portalURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		portalURL: strings.TrimRight(strings.TrimSpace(portalURL), "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) PortalURL() string {
	if c == nil {
		return ""
	}
	return c.portalURL
}

func (c *Client) FetchParallelPage(ctx context.Context, parallelID string) ([]byte, error) {
	if c == nil || c.portalURL == "" {
		return nil, fmt.Errorf("lksh: portal URL not configured")
	}
	parallelID = strings.ToLower(strings.Trim(strings.TrimSpace(parallelID), "/"))
	if parallelID == "" {
		return nil, fmt.Errorf("lksh: empty parallel id")
	}
	url := c.portalURL + "/" + parallelID + "/"
	return c.get(ctx, url)
}

func (c *Client) FetchPDF(ctx context.Context, pdfURL string) (io.ReadCloser, string, error) {
	if c == nil {
		return nil, "", fmt.Errorf("lksh: client not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pdfURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("lksh: fetch pdf: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", fmt.Errorf("lksh: fetch pdf: status %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/pdf"
	}
	return resp.Body, ct, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lksh: get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lksh: get %s: status %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lksh: read body: %w", err)
	}
	return body, nil
}
