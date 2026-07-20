package openai_test

import (
	"testing"

	"scainer/pkg/openai"
)

func TestLoadEnv_OK(t *testing.T) {
	t.Setenv(openai.EnvBaseURL, "https://example.test/v1")
	t.Setenv(openai.EnvUsername, "u")
	t.Setenv(openai.EnvPassword, "p")
	t.Setenv(openai.EnvTimeout, "5s")
	t.Setenv(openai.EnvModel, "llama3.2")
	t.Setenv(openai.EnvAPIKey, "")

	env, err := openai.LoadEnv()
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
	t.Setenv(openai.EnvBaseURL, "")
	t.Setenv(openai.EnvUsername, "u")
	t.Setenv(openai.EnvPassword, "p")
	t.Setenv(openai.EnvAPIKey, "")
	t.Setenv(openai.EnvTimeout, "")

	env, err := openai.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Client.BaseURL() != openai.DefaultBaseURL {
		t.Fatalf("base = %q", env.Client.BaseURL())
	}
}

func TestLoadEnv_MissingAuth(t *testing.T) {
	t.Setenv(openai.EnvBaseURL, openai.DefaultBaseURL)
	t.Setenv(openai.EnvUsername, "")
	t.Setenv(openai.EnvPassword, "")
	t.Setenv(openai.EnvAPIKey, "")

	_, err := openai.LoadEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}
