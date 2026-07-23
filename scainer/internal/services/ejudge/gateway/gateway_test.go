package gateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/teachers"
	"scainer/pkg/auth"
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
	if cred.LastSID != "" {
		existing.LastSID = cred.LastSID
	}
	if cred.APIKey != "" {
		existing.APIKey = cred.APIKey
	}
	existing.Login = cred.Login
	m.byLogin[cred.Login] = existing
	return nil
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

func TestGateway_ClientFor_Bootstrap(t *testing.T) {
	var createKeyCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/new-master":
			switch {
			case r.FormValue("action_2") == "Submit":
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
	if stored.APIKey != "generated-key" || stored.LastSID != "abc123" {
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
