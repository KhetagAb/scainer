package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrAPIKeyFailed = errors.New("ejudge: create API key failed")
	ErrAPIKeyParse  = errors.New("ejudge: API key token not found in response")
)

var apiKeyTokenRe = regexp.MustCompile(`(?is)API\s+key\s+token:\s*(?:</td><td[^>]*>\s*)?<tt>\s*([^<]+)\s*</tt>`)

func CreateAPIKey(ctx context.Context, baseURL string, session Session) (string, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	client := session.client(30 * time.Second)

	form := url.Values{}
	form.Set("SID", session.SID)
	form.Set("key_duration", "2592000") // 30 days
	form.Set("key_contest_id", "")
	form.Set("key_role", "6") // admin
	form.Set("action_311", "Submit")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/cgi-bin/new-master", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", ErrAPIKeyFailed
	}

	token, ok := parseAPIKeyToken(string(body))
	if !ok {
		return "", ErrAPIKeyParse
	}
	if token == "" {
		return "", ErrAPIKeyFailed
	}
	return token, nil
}

func IssueAPIKey(ctx context.Context, baseURL, login, password string) (string, error) {
	session, err := MasterSessionLogin(ctx, baseURL, login, password)
	if err != nil {
		return "", err
	}
	return CreateAPIKey(ctx, baseURL, session)
}

func parseAPIKeyToken(html string) (string, bool) {
	m := apiKeyTokenRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}
