package aiusage

import (
	"fmt"

	"scainer/internal/configs"
	"scainer/internal/domain"
	"scainer/internal/services/analyze/detect"
	"scainer/pkg/llm"
	"scainer/pkg/llm/openai"
)

func NewFromConfig(cfg configs.AIUsageConfig, openAI configs.OpenAIConfig) (detect.Detector[domain.ProblemParticipantUnit], error) {
	if !cfg.Enabled {
		return nil, nil
	}
	d, err := newTaskDetectorFromConfig(openAI)
	if err != nil {
		return nil, fmt.Errorf("aiusage: %w", err)
	}
	return d, nil
}

func newTaskDetectorFromConfig(cfg configs.OpenAIConfig) (*TaskDetector, error) {
	model, err := openAIModel(cfg)
	if err != nil {
		return nil, err
	}
	return NewTaskDetector(&Analyzer{Model: model}), nil
}

func openAIModel(cfg configs.OpenAIConfig) (llm.IntelligenceModel, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = llm.DefaultBaseURL
	}
	var (
		client *openai.Client
		err    error
	)
	if cfg.Username != "" || cfg.Password != "" {
		client, err = openai.New(baseURL, cfg.Username, cfg.Password, cfg.Timeout)
	} else if cfg.APIKey != "" {
		client, err = openai.NewWithAPIKey(baseURL, cfg.APIKey, cfg.Timeout)
	} else {
		return nil, fmt.Errorf("задайте openai.username/password или openai.api_key")
	}
	if err != nil {
		return nil, err
	}
	return llm.NewOpenAI(client, cfg.Model)
}
