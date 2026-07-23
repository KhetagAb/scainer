//go:build live

package ejudge

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"scainer/internal/configs"
	ejauth "scainer/pkg/ejudge/auth"
)

// LIVE=1 go test ./pkg/ejudge -tags=live -run TestLiveRun114Comments -v
func TestLiveRun114Comments(t *testing.T) {
	if os.Getenv("LIVE") == "" {
		t.Skip("set LIVE=1")
	}
	baseURL, password := teacherFromEnv(t, "khetag_dz")
	ctx := context.Background()

	session, err := ejauth.MasterSessionLogin(ctx, baseURL, "khetag_dz", password)
	if err != nil {
		t.Fatal(err)
	}
	apiKey, err := ejauth.CreateAPIKey(ctx, baseURL, session)
	if err != nil {
		t.Logf("create API key: %v (continuing with SID only)", err)
	}

	var bearerMsgs []RunMessage
	if apiKey != "" {
		bearerBody := fetchViewSource(t, ctx, baseURL, apiKey, "", 50506, 114)
		bearerMsgs, err = parseViewSourceComments(bearerBody)
		if err != nil {
			t.Fatalf("bearer parse: %v", err)
		}
		t.Logf("bearer HTML=%d bytes previous_section=%v comments=%d",
			len(bearerBody), strings.Contains(string(bearerBody), "Run comments for previous runs"), len(bearerMsgs))
	}

	sidBody := fetchViewSourceSID(t, ctx, session, 50506, 114)
	sidMsgs, err := parseViewSourceComments(sidBody)
	if err != nil {
		t.Fatalf("sid parse: %v", err)
	}

	t.Logf("sid HTML=%d bytes previous_section=%v comments=%d",
		len(sidBody), strings.Contains(string(sidBody), "Run comments for previous runs"), len(sidMsgs))
	for _, m := range sidMsgs {
		t.Logf("  [%d] %s @ %s", m.ClarID, m.From, m.Time.Format(time.RFC3339))
	}

	if len(sidMsgs) < 2 {
		if i := strings.Index(string(sidBody), "<title>"); i >= 0 {
			if j := strings.Index(string(sidBody[i:]), "</title>"); j > 0 {
				t.Logf("sid title: %s", string(sidBody[i:i+j+8]))
			}
		}
		t.Fatalf("expected >=2 comments via SID, got %d", len(sidMsgs))
	}
	if apiKey != "" && len(bearerMsgs) != len(sidMsgs) {
		t.Fatalf("bearer=%d sid=%d comments differ", len(bearerMsgs), len(sidMsgs))
	}
}

func fetchViewSource(t *testing.T, ctx context.Context, baseURL, apiKey, sid string, contestID, runID int) []byte {
	t.Helper()
	u, err := url.Parse(strings.TrimRight(baseURL, "/") + "/cgi-bin/new-master")
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("action", "36")
	q.Set("contest_id", strconv.Itoa(contestID))
	q.Set("run_id", strconv.Itoa(runID))
	if sid != "" {
		q.Set("SID", sid)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	client := http.DefaultClient
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer AQAA"+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP %d: %s", resp.StatusCode, truncateBody(body, 200))
	}
	return body
}

func fetchViewSourceSID(t *testing.T, ctx context.Context, session ejauth.Session, contestID, runID int) []byte {
	t.Helper()
	u, err := url.Parse(strings.TrimRight(session.BaseURL(), "/") + "/cgi-bin/new-master")
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("SID", session.SID)
	q.Set("action", "36")
	q.Set("run_id", strconv.Itoa(runID))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := session.HTTPClient(30 * time.Second).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP %d: %s", resp.StatusCode, truncateBody(body, 200))
	}
	return body
}

func teacherFromEnv(t *testing.T, login string) (baseURL, password string) {
	t.Helper()
	b, err := os.ReadFile("../../.env")
	if err != nil {
		t.Fatal(err)
	}
	var teachersRaw string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "EJUDGE_BASE_URL=") {
			baseURL = strings.TrimPrefix(line, "EJUDGE_BASE_URL=")
		}
		if strings.HasPrefix(line, "TEACHERS_PASSWORDS=") {
			teachersRaw = strings.TrimPrefix(line, "TEACHERS_PASSWORDS=")
		}
	}
	teachers, err := configs.ParseTeachersPasswords(strings.TrimSpace(teachersRaw))
	if err != nil {
		t.Fatal(err)
	}
	pw, ok := teachers[login]
	if !ok {
		t.Fatalf("login %q not in env", login)
	}
	return baseURL, pw
}
