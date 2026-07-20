package ejudge

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/lksh/scainer/internal/domain"
)

const (
	EnvBaseURL = "EJUDGE_BASE_URL"
	EnvAPIKey  = "EJUDGE_API_KEY"
	EnvTimeout = "EJUDGE_TIMEOUT"
)

// Точные ejudge lang_name → domain.Lang; остальные ловит guessLang.
var defaultLangMap = map[string]domain.Lang{
	"gcc": domain.LangCPP, "g++": domain.LangCPP, "clang++": domain.LangCPP,
	"python3": domain.LangPython, "python": domain.LangPython,
	"java": domain.LangJava,
	"go":   domain.LangGo,
}

type Config struct {
	ContestID int `yaml:"contest_id"`
}

type Env struct {
	Client  *Client
	LangMap map[string]domain.Lang
}

func LoadEnv() (*Env, error) {
	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		return nil, errors.New("ejudge: переменная окружения " + EnvBaseURL + " не задана")
	}
	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		return nil, errors.New("ejudge: переменная окружения " + EnvAPIKey + " не задана")
	}

	var timeout time.Duration
	if raw := os.Getenv(EnvTimeout); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, errors.New("ejudge: невалидный " + EnvTimeout + ": " + err.Error())
		}
		timeout = d
	}

	client, err := New(baseURL, apiKey, timeout)
	if err != nil {
		return nil, err
	}
	return &Env{Client: client, LangMap: copyLangMap()}, nil
}

func (e *Env) NormalizeLang(langName string) domain.Lang {
	if langName == "" {
		return ""
	}
	if e != nil && e.LangMap != nil {
		if v, ok := e.LangMap[langName]; ok {
			return v
		}
	} else if v, ok := defaultLangMap[langName]; ok {
		return v
	}
	return guessLang(langName)
}

func guessLang(langName string) domain.Lang {
	lower := strings.ToLower(langName)
	switch {
	case strings.Contains(lower, "clang++"),
		strings.HasPrefix(lower, "g++"),
		strings.HasPrefix(lower, "gcc"),
		strings.Contains(lower, "c++"):
		return domain.LangCPP
	case strings.HasPrefix(lower, "python"):
		return domain.LangPython
	case strings.HasPrefix(lower, "java"):
		return domain.LangJava
	case lower == "go" || strings.HasPrefix(lower, "golang"):
		return domain.LangGo
	case strings.Contains(lower, "javascript"), strings.HasPrefix(lower, "node"):
		return domain.LangJavaScript
	default:
		return domain.Lang(langName)
	}
}

func copyLangMap() map[string]domain.Lang {
	out := make(map[string]domain.Lang, len(defaultLangMap))
	for k, v := range defaultLangMap {
		out[k] = v
	}
	return out
}
