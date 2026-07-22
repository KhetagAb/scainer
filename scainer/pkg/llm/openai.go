package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	oaigen "scainer/generated/openai"
	"scainer/pkg/openai"
)

const promptAttempts = 3

type IntelligenceModel interface {
	Prompt(ctx context.Context, prompt string) (string, error)
}

type OpenAI struct {
	client *openai.Client
	model  string
}

var _ IntelligenceModel = (*OpenAI)(nil)

func NewOpenAI(client *openai.Client, model string) (*OpenAI, error) {
	if client == nil {
		return nil, fmt.Errorf("llm: openai client обязателен")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("llm: model обязателен")
	}
	return &OpenAI{client: client, model: model}, nil
}

func NewOpenAIFromEnv() (*OpenAI, error) {
	env, err := LoadEnv()
	if err != nil {
		return nil, err
	}
	return NewOpenAI(env.Client, env.Model)
}

func (m *OpenAI) Prompt(ctx context.Context, prompt string) (string, error) {
	var last error
	for attempt := 1; attempt <= promptAttempts; attempt++ {
		out, err := m.promptOnce(ctx, prompt)
		if err == nil {
			return out, nil
		}
		last = err
		if attempt == promptAttempts {
			break
		}
		delay := time.Duration(attempt) * 500 * time.Millisecond
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(delay):
		}
	}
	return "", last
}

func (m *OpenAI) promptOnce(ctx context.Context, prompt string) (string, error) {
	resp, err := m.client.CreateChatCompletionWithResponse(ctx, oaigen.ChatCompletionRequest{
		Model: m.model,
		Messages: []oaigen.ChatMessage{{
			Role:    "user",
			Content: prompt,
		}},
	})
	if err != nil {
		return "", fmt.Errorf("llm: chat completion: %w", err)
	}
	if err := openai.EnsureOK(resp.StatusCode(), resp.Body); err != nil {
		return "", err
	}
	if resp.JSON200 == nil || resp.JSON200.Choices == nil || len(*resp.JSON200.Choices) == 0 {
		return "", fmt.Errorf("llm: пустой ответ модели")
	}
	msg := (*resp.JSON200.Choices)[0].Message
	if msg == nil {
		return "", fmt.Errorf("llm: пустое сообщение ассистента")
	}
	return msg.Content, nil
}
