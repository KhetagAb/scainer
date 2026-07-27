package lksh

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func FindLessonByContestID(lessons []Lesson, contestID string) (Lesson, error) {
	contestID = strings.TrimSpace(contestID)
	if contestID == "" {
		return Lesson{}, fmt.Errorf("lksh: empty contest id")
	}
	for _, lesson := range lessons {
		id, ok := ParseEjudgeContestID(lesson.ContestURL)
		if !ok {
			continue
		}
		if id == contestID {
			return lesson, nil
		}
	}
	return Lesson{}, fmt.Errorf("lksh: lesson for contest %s not found", contestID)
}

func ParseEjudgeContestID(contestURL string) (string, bool) {
	contestURL = strings.TrimSpace(contestURL)
	if contestURL == "" {
		return "", false
	}
	if !strings.Contains(strings.ToLower(contestURL), "ejudge") {
		return "", false
	}
	u, err := url.Parse(contestURL)
	if err != nil {
		return "", false
	}
	id := strings.TrimSpace(u.Query().Get("contest_id"))
	if id == "" {
		return "", false
	}
	if _, err := strconv.Atoi(id); err != nil {
		return "", false
	}
	return id, true
}

func BuildStatementPDFURL(portalURL, parallelID, baseURL string, lesson Lesson) (string, error) {
	statementsURL := strings.TrimSpace(lesson.StatementsURL)
	if statementsURL == "" {
		return "", fmt.Errorf("lksh: empty statements_url")
	}
	portalURL = strings.TrimRight(strings.TrimSpace(portalURL), "/")
	parallelID = strings.ToLower(strings.Trim(strings.TrimSpace(parallelID), "/"))
	if portalURL == "" || parallelID == "" {
		return "", fmt.Errorf("lksh: portal url or parallel id missing")
	}

	primary := portalURL + "/" + parallelID + "/" + strings.TrimLeft(statementsURL, "/")
	if u, err := url.Parse(primary); err == nil && u.Scheme != "" && u.Host != "" {
		return u.String(), nil
	}

	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "/" + parallelID + "/"
	}
	ref, err := url.Parse(portalURL + "/" + strings.Trim(base, "/") + "/")
	if err != nil {
		return "", fmt.Errorf("lksh: resolve base url: %w", err)
	}
	rel, err := url.Parse(statementsURL)
	if err != nil {
		return "", fmt.Errorf("lksh: resolve statements url: %w", err)
	}
	return ref.ResolveReference(rel).String(), nil
}

func StatementPDFURLFromPage(portalURL, parallelID, contestID string, pageHTML []byte) (string, error) {
	cfg, err := ExtractRawConfig(pageHTML)
	if err != nil {
		return "", err
	}
	lessons := CollectLessons(cfg.Layout)
	lesson, err := FindLessonByContestID(lessons, contestID)
	if err != nil {
		return "", err
	}
	return BuildStatementPDFURL(portalURL, parallelID, cfg.BaseURL, lesson)
}
