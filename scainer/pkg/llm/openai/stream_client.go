package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionStreamRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// StreamChatCompletion открывает поток chat/completions с stream=true.
func (c *Client) StreamChatCompletion(ctx context.Context, model, prompt string) (<-chan string, <-chan error) {
	out := make(chan string, 32)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		body, err := json.Marshal(chatCompletionStreamRequest{
			Model: model,
			Messages: []chatMessage{{
				Role:    "user",
				Content: prompt,
			}},
			Stream: true,
		})
		if err != nil {
			errCh <- fmt.Errorf("openai: marshal stream request: %w", err)
			return
		}

		req, err := c.NewRequest(ctx, http.MethodPost, "/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// Для стрима не используем общий Timeout клиента — только ctx.
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			errCh <- fmt.Errorf("openai: stream request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(resp.Body)
			errCh <- EnsureOK(resp.StatusCode, respBody)
			return
		}

		if err := ParseChatCompletionSSE(resp.Body, func(delta string) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out <- delta:
				return nil
			}
		}); err != nil {
			errCh <- err
		}
	}()

	return out, errCh
}
