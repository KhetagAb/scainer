package ejudge

import (
	"strings"

	"scainer/internal/domain"
)

var defaultLangMap = map[string]domain.Lang{
	"gcc": domain.LangCPP, "g++": domain.LangCPP, "clang++": domain.LangCPP,
	"python3": domain.LangPython, "python": domain.LangPython,
	"java": domain.LangJava,
	"go":   domain.LangGo,
}

func NormalizeLang(langName string) domain.Lang {
	if langName == "" {
		return ""
	}
	if v, ok := defaultLangMap[langName]; ok {
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
