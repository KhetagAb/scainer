package ejudge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRunMessages_ViewSourceHTML(t *testing.T) {
	fixture := readTestdata(t, "view_source_run_comments.html")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("action") != "36" {
			t.Fatalf("action=%s", q.Get("action"))
		}
		if q.Get("contest_id") != "50504" || q.Get("run_id") != "210" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer AQAAtok" {
			t.Fatalf("auth=%q", got)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	c, err := New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := c.RunMessages(context.Background(), 50504, 210)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("len=%d %+v", len(msgs), msgs)
	}
	m := msgs[0]
	if m.From != "Хетаг Дзестелов" {
		t.Fatalf("from=%q", m.From)
	}
	if m.Text != "еуые" {
		t.Fatalf("text=%q", m.Text)
	}
	want := time.Date(2026, 7, 22, 18, 11, 36, 0, time.FixedZone("MSK", 3*3600))
	if !m.Time.Equal(want) {
		t.Fatalf("time=%v want=%v", m.Time, want)
	}
}

func TestParseViewSourceComments_EmptySection(t *testing.T) {
	msgs, err := parseViewSourceComments([]byte(`<html><h2>Change status</h2></html>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Fatalf("len=%d", len(msgs))
	}
}

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	path := filepath.Join(filepath.Dir(file), "testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
