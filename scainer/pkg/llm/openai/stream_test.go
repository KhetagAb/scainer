package openai_test

import (
	"strings"
	"testing"

	"scainer/pkg/llm/openai"
)

func TestParseChatCompletionSSE(t *testing.T) {
	var got strings.Builder
	err := openai.ParseChatCompletionSSE(
		strings.NewReader(string(openai.FormatChatCompletionSSE("hel", "lo"))),
		func(delta string) error {
			got.WriteString(delta)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "hello" {
		t.Fatalf("got %q", got.String())
	}
}

func TestParseChatCompletionSSE_ignoresDoneSentinel(t *testing.T) {
	body := "data: [DONE]\n\n"
	err := openai.ParseChatCompletionSSE(strings.NewReader(body), func(string) error {
		t.Fatal("unexpected delta")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
