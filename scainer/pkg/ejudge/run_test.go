package ejudge_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"scainer/pkg/ejudge"
)

func TestRunStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" || r.URL.Query().Get("action") != "run-status-json" {
			t.Fatalf("unexpected %s %s", r.URL.Path, r.URL.RawQuery)
		}
		if r.URL.Query().Get("run_id") != "7" {
			t.Fatalf("run_id=%s", r.URL.Query().Get("run_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"run":{"run_id":7,"status":17,"status_str":"RJ"}}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	info, err := c.RunStatus(context.Background(), 42, 7)
	if err != nil {
		t.Fatal(err)
	}
	if info.RunID != 7 || info.Status != 17 || info.Verdict != ejudge.VerdictRJ {
		t.Fatalf("got %+v", info)
	}
}

func TestSendRunComment(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/cgi-bin/new-master" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SendRunComment(context.Background(), 42, 7, "fix"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotBody, "action=64") || !strings.Contains(gotBody, "msg_text=fix") {
		t.Fatalf("body=%q", gotBody)
	}
}

func TestChangeRunStatus(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ChangeRunStatus(context.Background(), 42, 7, ejudge.VerdictOK); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotBody, "action=67") || !strings.Contains(gotBody, "status=0") {
		t.Fatalf("body=%q", gotBody)
	}
}

func TestChangeRunStatus_EmptyJSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// ejudge часто отвечает 200 + пустое тело на action=67
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.ChangeRunStatus(context.Background(), 42, 7, ejudge.VerdictOK); err != nil {
		t.Fatal(err)
	}
}

func TestChangeRunStatus_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":false,"error":{"message":"denied"}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = c.ChangeRunStatus(context.Background(), 42, 7, ejudge.VerdictRJ)
	if err == nil || !strings.Contains(err.Error(), "denied") {
		t.Fatalf("err=%v", err)
	}
}
