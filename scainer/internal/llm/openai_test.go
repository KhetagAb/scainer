package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	oaigen "scainer/generated/openai"
	"scainer/internal/llm"
	"scainer/pkg/openai"
)

func TestOpenAI_Prompt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req oaigen.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "m" || len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Fatalf("req = %+v", req)
		}
		if req.Messages[0].Content != "hello" {
			t.Fatalf("content = %q", req.Messages[0].Content)
		}
		msg := oaigen.ChatMessage{Role: "assistant", Content: "world"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oaigen.ChatCompletionResponse{
			Choices: &[]oaigen.ChatCompletionChoice{{Message: &msg}},
		})
	}))
	t.Cleanup(srv.Close)

	client, err := openai.NewWithAPIKey(srv.URL+"/v1", "k", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	model, err := llm.NewOpenAI(client, "m")
	if err != nil {
		t.Fatal(err)
	}
	got, err := model.Prompt(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "world" {
		t.Fatalf("got %q", got)
	}
}

func TestNewOpenAI_RequiresModel(t *testing.T) {
	client, err := openai.NewWithAPIKey("http://example.test/v1", "k", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = llm.NewOpenAI(client, "  ")
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("err = %v", err)
	}
}
