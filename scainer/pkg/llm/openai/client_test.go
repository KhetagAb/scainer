package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	oaigen "scainer/generated/openai"
	"scainer/pkg/llm/openai"
)

func TestListModels_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "u" || pass != "p" {
			t.Errorf("basic auth = %q/%q ok=%v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oaigen.ModelList{
			Object: "list",
			Data:   []oaigen.Model{{Id: "llama3.2", Object: "model"}},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := openai.New(srv.URL+"/v1", "u", "p", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.ListModelsWithResponse(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := openai.EnsureOK(resp.StatusCode(), resp.Body); err != nil {
		t.Fatal(err)
	}
	if resp.JSON200 == nil || len(resp.JSON200.Data) != 1 || resp.JSON200.Data[0].Id != "llama3.2" {
		t.Fatalf("resp = %+v body=%s", resp.JSON200, string(resp.Body))
	}
}

func TestCreateChatCompletion_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		var req oaigen.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "m" || len(req.Messages) != 1 || req.Messages[0].Content != "ping" {
			t.Fatalf("req = %+v", req)
		}
		msg := oaigen.ChatMessage{Role: "assistant", Content: "pong"}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oaigen.ChatCompletionResponse{
			Id:      ptr("chatcmpl-1"),
			Model:   ptr("m"),
			Choices: &[]oaigen.ChatCompletionChoice{{Message: &msg, FinishReason: ptr("stop")}},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := openai.NewWithAPIKey(srv.URL+"/v1", "secret", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.CreateChatCompletionWithResponse(context.Background(), oaigen.ChatCompletionRequest{
		Model:    "m",
		Messages: []oaigen.ChatMessage{{Role: "user", Content: "ping"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := openai.EnsureOK(resp.StatusCode(), resp.Body); err != nil {
		t.Fatal(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Choices == nil || len(*resp.JSON200.Choices) != 1 {
		t.Fatalf("resp = %+v body=%s", resp.JSON200, string(resp.Body))
	}
	got := (*resp.JSON200.Choices)[0].Message
	if got == nil || got.Content != "pong" {
		t.Fatalf("message = %+v", got)
	}
}

func TestEnsureOK_Error(t *testing.T) {
	err := openai.EnsureOK(http.StatusUnauthorized, nil)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v", err)
	}
}

func ptr[T any](v T) *T { return &v }
