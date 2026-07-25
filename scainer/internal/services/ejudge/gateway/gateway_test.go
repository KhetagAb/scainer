package gateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/teachers"
	"scainer/pkg/auth"
	ejauth "scainer/pkg/ejudge/auth"
)

type memTeachers struct {
	rec teachers.Record
	ok  bool
}

func (m *memTeachers) Get(_ context.Context, login string) (teachers.Record, bool, error) {
	if !m.ok || m.rec.Login != login {
		return teachers.Record{}, false, nil
	}
	return m.rec, true, nil
}

type memCreds struct {
	byLogin map[string]gateway.Credentials
}

func (m *memCreds) Get(_ context.Context, login string) (gateway.Credentials, bool, error) {
	c, ok := m.byLogin[login]
	return c, ok, nil
}

func (m *memCreds) Upsert(_ context.Context, cred gateway.Credentials) error {
	existing := m.byLogin[cred.Login]
	if cred.APIKey != "" {
		existing.APIKey = cred.APIKey
	}
	existing.Login = cred.Login
	m.byLogin[cred.Login] = existing
	return nil
}

func masterLoginHandler(t *testing.T, loginCalls *int, contestID int, sid string) http.HandlerFunc {
	t.Helper()
	wantContest := strconv.Itoa(contestID)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("action_2") == "Submit" {
			*loginCalls++
			if got := r.FormValue("contest_id"); got != wantContest {
				t.Errorf("contest_id=%q want %q", got, wantContest)
			}
			_, _ = w.Write([]byte(`<a href="?SID=` + sid + `">home</a>`))
			return
		}
		if r.FormValue("action_311") == "Submit" {
			_, _ = w.Write([]byte(`<tr><td>API key token:</td><td><tt>generated-key</tt></td></tr>`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func TestGateway_ClientFor_ExistingAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	gw := gateway.New(nil, &memCreds{
		byLogin: map[string]gateway.Credentials{
			"alice": {Login: "alice", APIKey: "tok"},
		},
	}, srv.URL, time.Second)

	ctx := auth.WithLogin(context.Background(), "alice")
	client, err := gw.ClientFor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
}

func TestGateway_ClientFor_ProvisionAPIKey(t *testing.T) {
	var createKeyCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/new-master":
			switch {
			case r.FormValue("action_2") == "Submit":
				if got := r.FormValue("contest_id"); got != strconv.Itoa(ejauth.MasterLoginContestID) {
					t.Fatalf("contest_id=%q", got)
				}
				_, _ = w.Write([]byte(`<a href="?SID=abc123">home</a>`))
			case r.FormValue("action_311") == "Submit":
				createKeyCalls++
				_, _ = w.Write([]byte(`<tr><td>API key token:</td><td><tt>generated-key</tt></td></tr>`))
			default:
				http.NotFound(w, r)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	creds := &memCreds{byLogin: map[string]gateway.Credentials{}}
	gw := gateway.New(&memTeachers{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	}, creds, srv.URL, time.Second)

	ctx := auth.WithLogin(context.Background(), "alice")
	client, err := gw.ClientFor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
	if createKeyCalls != 1 {
		t.Fatalf("createKeyCalls=%d", createKeyCalls)
	}
	stored := creds.byLogin["alice"]
	if stored.APIKey != "generated-key" {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestGateway_ClientFor_NoLogin(t *testing.T) {
	gw := gateway.New(nil, &memCreds{byLogin: map[string]gateway.Credentials{}}, "http://example", time.Second)
	_, err := gw.ClientFor(context.Background())
	if !errors.Is(err, gateway.ErrNoLogin) {
		t.Fatalf("err=%v", err)
	}
}

func TestEnsureAPIKey_SingleMasterLogin(t *testing.T) {
	var loginCalls int
	srv := httptest.NewServer(masterLoginHandler(t, &loginCalls, ejauth.MasterLoginContestID, "abc123"))
	t.Cleanup(srv.Close)

	creds := &memCreds{byLogin: map[string]gateway.Credentials{}}
	gw := gateway.New(&memTeachers{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	}, creds, srv.URL, time.Second)

	ctx := auth.WithLogin(context.Background(), "alice")
	if err := gw.EnsureAPIKey(ctx); err != nil {
		t.Fatal(err)
	}
	if loginCalls != 1 {
		t.Fatalf("loginCalls=%d want 1 on cold start", loginCalls)
	}
	stored := creds.byLogin["alice"]
	if stored.APIKey != "generated-key" {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestEnsureAPIKey_NoOpWhenAPIKeyExists(t *testing.T) {
	var loginCalls int
	srv := httptest.NewServer(masterLoginHandler(t, &loginCalls, ejauth.MasterLoginContestID, "6d670df1467ca890"))
	t.Cleanup(srv.Close)

	creds := &memCreds{
		byLogin: map[string]gateway.Credentials{
			"alice": {Login: "alice", APIKey: "tok"},
		},
	}
	gw := gateway.New(&memTeachers{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	}, creds, srv.URL, time.Second)

	ctx := auth.WithLogin(context.Background(), "alice")
	if err := gw.EnsureAPIKey(ctx); err != nil {
		t.Fatal(err)
	}
	if loginCalls != 0 {
		t.Fatalf("loginCalls=%d want 0", loginCalls)
	}
}

func TestGateway_BrowserLogin_OK(t *testing.T) {
	const contestID = 50506
	gw := gateway.New(&memTeachers{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	}, &memCreds{
		byLogin: map[string]gateway.Credentials{
			"alice": {Login: "alice", APIKey: "tok"},
		},
	}, "https://ejudge.example", time.Second)

	ctx := auth.WithLogin(context.Background(), "alice")
	creds, ok, err := gw.BrowserLogin(ctx, contestID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || creds.BaseURL != "https://ejudge.example" {
		t.Fatalf("ok=%v creds=%+v", ok, creds)
	}
	if creds.Login != "alice" || creds.Password != "secret" || creds.ContestID != contestID {
		t.Fatalf("creds=%+v", creds)
	}
}

func TestGateway_BrowserLogin_NoBaseURL(t *testing.T) {
	gw := gateway.New(nil, &memCreds{byLogin: map[string]gateway.Credentials{}}, "", time.Second)
	_, ok, err := gw.BrowserLogin(auth.WithLogin(context.Background(), "alice"), 50506)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected ok=false when base URL empty")
	}
}

func TestGateway_BrowserLogin_NoContestID(t *testing.T) {
	gw := gateway.New(nil, &memCreds{byLogin: map[string]gateway.Credentials{}}, "http://example", time.Second)
	_, ok, err := gw.BrowserLogin(auth.WithLogin(context.Background(), "alice"), 0)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected ok=false when contest_id missing")
	}
}
