package llm_test

import (
	"testing"

	"scainer/pkg/llm"
)

func TestLoadEnv_OK(t *testing.T) {
	t.Setenv(llm.EnvBaseURL, "https://example.test/v1")
	t.Setenv(llm.EnvUsername, "u")
	t.Setenv(llm.EnvPassword, "p")
	t.Setenv(llm.EnvTimeout, "5s")
	t.Setenv(llm.EnvModel, "llama3.2")

	env, err := llm.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Client.Timeout().String() != "5s" {
		t.Fatalf("timeout = %v", env.Client.Timeout())
	}
	if env.Client.BaseURL() != "https://example.test/v1" {
		t.Fatalf("base = %q", env.Client.BaseURL())
	}
	if env.Model != "llama3.2" {
		t.Fatalf("model = %q", env.Model)
	}
}

func TestLoadEnv_DefaultBaseURL(t *testing.T) {
	t.Setenv(llm.EnvBaseURL, "")
	t.Setenv(llm.EnvUsername, "u")
	t.Setenv(llm.EnvPassword, "p")
	t.Setenv(llm.EnvTimeout, "")

	env, err := llm.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Client.BaseURL() != llm.DefaultBaseURL {
		t.Fatalf("base = %q", env.Client.BaseURL())
	}
}

func TestLoadEnv_MissingAuth(t *testing.T) {
	t.Setenv(llm.EnvBaseURL, llm.DefaultBaseURL)
	t.Setenv(llm.EnvUsername, "")
	t.Setenv(llm.EnvPassword, "")
	t.Setenv(llm.EnvAPIKey, "")

	_, err := llm.LoadEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadEnv_APIKey(t *testing.T) {
	t.Setenv(llm.EnvBaseURL, "https://llm.api.cloud.yandex.net/v1")
	t.Setenv(llm.EnvUsername, "")
	t.Setenv(llm.EnvPassword, "")
	t.Setenv(llm.EnvAPIKey, "secret")
	t.Setenv(llm.EnvModel, "gpt://folder/yandexgpt/latest")

	env, err := llm.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Client.BaseURL() != "https://llm.api.cloud.yandex.net/v1" {
		t.Fatalf("base = %q", env.Client.BaseURL())
	}
}
