package auth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	placeholderSID       = "0000000000000000"
	roleAdmin            = "6"
	MasterLoginContestID = 90000
)

var (
	ErrSIDNotFound = errors.New("ejudge: SID not found in login response")
	ErrInvalidSID  = errors.New("ejudge: invalid SID")
)

var sidParamRe = regexp.MustCompile(`(?i)SID=([0-9a-fA-F]+)`)

type Session struct {
	SID  string
	jar  http.CookieJar
	base string
}

// MasterSessionLogin — логин в new-master для получения SID (contest_id=90000).
func MasterSessionLogin(ctx context.Context, baseURL, login, password string) (Session, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	jar, err := cookiejar.New(nil)
	if err != nil {
		return Session{}, err
	}
	client := &http.Client{Timeout: 30 * time.Second, Jar: jar}

	form := url.Values{}
	form.Set("login", login)
	form.Set("password", password)
	form.Set("contest_id", strconv.Itoa(MasterLoginContestID))
	form.Set("role", roleAdmin)
	form.Set("action_2", "Submit")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/cgi-bin/new-master", strings.NewReader(form.Encode()))
	if err != nil {
		return Session{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return Session{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Session{}, err
	}

	sid := parseSID(string(body))
	if sid == "" {
		return Session{}, ErrSIDNotFound
	}
	if sid == placeholderSID {
		return Session{}, ErrInvalidSID
	}
	return Session{SID: sid, jar: jar, base: baseURL}, nil
}

func parseSID(s string) string {
	m := sidParamRe.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func (s Session) HTTPClient(timeout time.Duration) *http.Client {
	return s.client(timeout)
}

func (s Session) BaseURL() string { return s.base }

func (s Session) client(timeout time.Duration) *http.Client {
	if s.jar != nil {
		return &http.Client{Timeout: timeout, Jar: s.jar}
	}
	jar, _ := cookiejar.New(nil)
	if s.base != "" && s.SID != "" {
		if u, err := url.Parse(s.base); err == nil {
			jar.SetCookies(u, []*http.Cookie{{Name: "SID", Value: s.SID, Path: "/"}})
		}
	}
	return &http.Client{Timeout: timeout, Jar: jar}
}
