package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"scainer/pkg/ejudge/auth"
)

func TestMasterSessionLogin_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" {
			http.NotFound(w, r)
			return
		}
		if got := r.FormValue("contest_id"); got != "90000" {
			t.Fatalf("contest_id=%q", got)
		}
		if got := r.FormValue("role"); got != "6" {
			t.Fatalf("role=%q", got)
		}
		_, _ = w.Write([]byte(`<a href="?SID=abc123def456">home</a>`))
	}))
	t.Cleanup(srv.Close)

	session, err := auth.MasterSessionLogin(context.Background(), srv.URL, "alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if session.SID != "abc123def456" {
		t.Fatalf("SID=%q", session.SID)
	}
}

func TestMasterSessionLogin_PlaceholderRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`SID=0000000000000000`))
	}))
	t.Cleanup(srv.Close)

	_, err := auth.MasterSessionLogin(context.Background(), srv.URL, "alice", "wrong")
	if err == nil || !strings.Contains(err.Error(), "invalid SID") {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateAPIKey_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" {
			http.NotFound(w, r)
			return
		}
		if got := r.FormValue("SID"); got != "sid42" {
			t.Fatalf("SID=%q", got)
		}
		if got := r.FormValue("action_311"); got != "Submit" {
			t.Fatalf("action_311=%q", got)
		}
		if got := r.FormValue("key_duration"); got != "2592000" {
			t.Fatalf("key_duration=%q", got)
		}
		if got := r.FormValue("key_role"); got != "6" {
			t.Fatalf("key_role=%q", got)
		}
		_, _ = w.Write([]byte(`<tr><td>API key token:</td><td><tt>my-api-token</tt></td></tr>`))
	}))
	t.Cleanup(srv.Close)

	token, err := auth.CreateAPIKey(context.Background(), srv.URL, auth.Session{SID: "sid42"})
	if err != nil {
		t.Fatal(err)
	}
	if token != "my-api-token" {
		t.Fatalf("token=%q", token)
	}
}
