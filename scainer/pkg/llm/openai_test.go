package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	oaigen "scainer/generated/openai"
	"scainer/pkg/llm"
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

func TestOpenAI_PromptRetriesTransientHTTPError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid character '\\x00' looking for beginning of value"}}`))
			return
		}
		msg := oaigen.ChatMessage{Role: "assistant", Content: "ok"}
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
	if got != "ok" {
		t.Fatalf("got %q", got)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", calls.Load())
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
