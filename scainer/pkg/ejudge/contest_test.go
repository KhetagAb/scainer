package ejudge_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lksh/scainer/pkg/ejudge"
)

func TestContestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/master" || r.URL.Query().Get("action") != "contest-status-json" {
			t.Fatalf("unexpected %s %s", r.URL.Path, r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("missing auth")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"contest":{"id":50601,"name":"День 01"}}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	info, err := c.ContestStatus(context.Background(), 50601)
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != 50601 || info.Name != "День 01" {
		t.Fatalf("got %+v", info)
	}
}
