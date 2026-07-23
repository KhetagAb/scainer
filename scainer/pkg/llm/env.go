package llm

import (
	"errors"
	"os"
	"time"

	"scainer/pkg/llm/openai"
)

const (
	EnvBaseURL  = "OPENAI_BASE_URL"
	EnvUsername = "OPENAI_USERNAME"
	EnvPassword = "OPENAI_PASSWORD"
	EnvAPIKey   = "OPENAI_API_KEY"
	EnvTimeout  = "OPENAI_TIMEOUT"
	EnvModel    = "OPENAI_MODEL"

	DefaultBaseURL = "https://ollamaapi.n.icpc.live/v1"
)

type Env struct {
	Client *openai.Client
	Model  string
}

func LoadEnv() (*Env, error) {
	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	user := os.Getenv(EnvUsername)
	pass := os.Getenv(EnvPassword)
	apiKey := os.Getenv(EnvAPIKey)
	if user == "" && pass == "" && apiKey == "" {
		return nil, errors.New("llm: задайте " + EnvUsername + "/" + EnvPassword + " или " + EnvAPIKey)
	}

	var timeout time.Duration
	if raw := os.Getenv(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, errors.New("llm: невалидный " + EnvTimeout + ": " + err.Error())
		}
		timeout = d
	}

	var (
		client *openai.Client
		err    error
	)
	if user != "" || pass != "" {
		client, err = openai.New(baseURL, user, pass, timeout)
	} else {
		client, err = openai.NewWithAPIKey(baseURL, apiKey, timeout)
	}
	if err != nil {
		return nil, err
	}
	return &Env{Client: client, Model: os.Getenv(EnvModel)}, nil
}
