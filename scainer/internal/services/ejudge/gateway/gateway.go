package gateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"scainer/internal/services/teachers"
	"scainer/pkg/auth"
	"scainer/pkg/ejudge"
	ejauth "scainer/pkg/ejudge/auth"
)

type teacherRepository interface {
	Get(ctx context.Context, login string) (teachers.Record, bool, error)
}

type CredentialsRepository interface {
	Get(ctx context.Context, login string) (Credentials, bool, error)
	Upsert(ctx context.Context, cred Credentials) error
}

type Gateway struct {
	teachers    teacherRepository
	credentials CredentialsRepository
	
	baseURL     string
	timeout     time.Duration

	bootstrapMu sync.Mutex
}

func New(teachers teacherRepository, credentials CredentialsRepository, baseURL string, timeout time.Duration) *Gateway {
	return &Gateway{
		teachers:    teachers,
		credentials: credentials,
		baseURL:     baseURL,
		timeout:     timeout,
	}
}

func (g *Gateway) ClientFor(ctx context.Context) (*ejudge.Client, error) {
	if g.baseURL == "" {
		return nil, fmt.Errorf("ejudge gateway: base URL not configured")
	}

	login, ok := auth.LoginFrom(ctx)
	if !ok || login == "" {
		return nil, ErrNoLogin
	}

	client, err := g.clientWithAPIKey(ctx, login)
	if err == nil {
		return client, nil
	}
	if !errors.Is(err, errAPIKeyMissing) {
		return nil, err
	}

	g.bootstrapMu.Lock()
	defer g.bootstrapMu.Unlock()

	client, err = g.clientWithAPIKey(ctx, login)
	if err == nil {
		return client, nil
	}
	if !errors.Is(err, errAPIKeyMissing) {
		return nil, err
	}

	if err := g.bootstrap(ctx, login); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBootstrap, err)
	}

	return g.clientWithAPIKey(ctx, login)
}

var errAPIKeyMissing = errors.New("ejudge gateway: api key missing")

func (g *Gateway) clientWithAPIKey(ctx context.Context, login string) (*ejudge.Client, error) {
	cred, found, err := g.credentials.Get(ctx, login)
	if err != nil {
		return nil, err
	}
	if !found || cred.APIKey == "" {
		return nil, errAPIKeyMissing
	}
	return ejudge.New(g.baseURL, cred.APIKey, g.timeout)
}

func (g *Gateway) bootstrap(ctx context.Context, login string) error {
	rec, ok, err := g.teachers.Get(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTeacherNotFound
	}

	session, err := ejauth.MasterSessionLogin(ctx, g.baseURL, login, rec.Password)
	if err != nil {
		return err
	}

	if err := g.credentials.Upsert(ctx, Credentials{
		Login:   login,
		LastSID: session.SID,
	}); err != nil {
		return err
	}

	apiKey, err := ejauth.CreateAPIKey(ctx, g.baseURL, session)
	if err != nil {
		return err
	}

	return g.credentials.Upsert(ctx, Credentials{
		Login:  login,
		APIKey: apiKey,
	})
}
