package openai

import (
	"errors"
	"os"
	"time"
)

const (
	EnvBaseURL  = "OPENAI_BASE_URL"
	EnvUsername = "OPENAI_USERNAME"
	EnvPassword = "OPENAI_PASSWORD"
	EnvAPIKey   = "OPENAI_API_KEY"
	EnvTimeout  = "OPENAI_TIMEOUT"
	EnvModel    = "OPENAI_MODEL"
)

type Env struct {
	Client *Client
	Model  string
}

func LoadEnv() (*Env, error) {
	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		return nil, errors.New("openai: задайте " + EnvBaseURL)
	}

	user := os.Getenv(EnvUsername)
	pass := os.Getenv(EnvPassword)
	apiKey := os.Getenv(EnvAPIKey)
	if user == "" && pass == "" && apiKey == "" {
		return nil, errors.New("openai: задайте " + EnvUsername + "/" + EnvPassword + " или " + EnvAPIKey)
	}

	var timeout time.Duration
	if raw := os.Getenv(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, errors.New("openai: невалидный " + EnvTimeout + ": " + err.Error())
		}
		timeout = d
	}

	var (
		client *Client
		err    error
	)
	if user != "" || pass != "" {
		client, err = New(baseURL, user, pass, timeout)
	} else {
		client, err = NewWithAPIKey(baseURL, apiKey, timeout)
	}
	if err != nil {
		return nil, err
	}
	return &Env{Client: client, Model: os.Getenv(EnvModel)}, nil
}
