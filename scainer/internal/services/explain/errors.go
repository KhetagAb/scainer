package explain

import "errors"

var (
	ErrNotFound      = errors.New("problem statement not found")
	ErrNotConfigured = errors.New("ai not configured")
	ErrLLM           = errors.New("llm request failed")
)
