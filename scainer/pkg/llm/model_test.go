package llm_test

import (
	"testing"

	"scainer/pkg/llm"
)

func TestUnconfigured_ReturnsErrNotConfigured(t *testing.T) {
	_, err := llm.Unconfigured{}.Prompt(t.Context(), "x")
	if err != llm.ErrNotConfigured {
		t.Fatalf("err=%v", err)
	}
	_, err = llm.Unconfigured{}.PromptStream(t.Context(), "x")
	if err != llm.ErrNotConfigured {
		t.Fatalf("stream err=%v", err)
	}
}
