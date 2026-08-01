package llm

import (
	"context"
	"errors"
)

var ErrNotConfigured = errors.New("llm: not configured")

type Model interface {
	Prompt(ctx context.Context, prompt string) (string, error)
	PromptStream(ctx context.Context, prompt string) (<-chan StreamChunk, error)
	ModelName() string
}

type Unconfigured struct{}

var _ Model = Unconfigured{}

func (Unconfigured) Prompt(context.Context, string) (string, error) {
	return "", ErrNotConfigured
}

func (Unconfigured) PromptStream(context.Context, string) (<-chan StreamChunk, error) {
	return nil, ErrNotConfigured
}

func (Unconfigured) ModelName() string { return "" }
