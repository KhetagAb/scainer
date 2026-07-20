package ejudge_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ejgen "scainer/generated/ejudge"
	"scainer/pkg/ejudge"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestListRuns_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("json") != "1" || q.Get("action") != "list-runs-json" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		writeJSON(w, map[string]any{
			"ok": true,
			"result": map[string]any{
				"runs": []map[string]any{
					{"run_id": 1, "user_name": "alice", "prob_internal_name": "A"},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	first := 0
	resp, err := c.ListRunsWithResponse(context.Background(), ejudge.ListRunsParams(42, &first, nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.JSON200 == nil {
		t.Fatalf("status=%d body=%s", resp.StatusCode(), string(resp.Body))
	}
	if err := ejudge.EnsureOK(resp.JSON200.Ok, resp.JSON200.Error); err != nil {
		t.Fatal(err)
	}
	runs := resp.JSON200.Result.Runs
	if runs == nil || len(*runs) != 1 || (*runs)[0].UserName == nil || *(*runs)[0].UserName != "alice" {
		t.Fatalf("runs = %+v", runs)
	}
}

func TestEnsureOK_APIError(t *testing.T) {
	msg := "permission denied"
	err := ejudge.EnsureOK(boolPtr(false), &ejgen.Error{Message: &msg})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("error = %v", err)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{"runs": []any{}}})
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "secret-token", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.ListRunsWithResponse(context.Background(), ejudge.ListRunsParams(1, nil, nil))
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer AQAAsecret-token" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
}

func TestDownloadRun_Raw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/master" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("json") != "1" || q.Get("action") != "download-run" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte("#include <stdio.h>\n"))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.DownloadRunWithResponse(context.Background(), ejudge.DownloadRunParams(1, 7))
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Body) != "#include <stdio.h>\n" {
		t.Fatalf("body = %q", resp.Body)
	}
}

func TestNew_DefaultTimeout(t *testing.T) {
	c, err := ejudge.New("https://ejudge.lksh.ru", "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	if c.Timeout() != 15*time.Second {
		t.Fatalf("timeout = %v", c.Timeout())
	}
}

func boolPtr(b bool) *bool { return &b }
