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
	"scainer/pkg/ejudge/servecontrol"
)

type teacherRepository interface {
	Get(ctx context.Context, login string) (teachers.Record, bool, error)
}

// BrowserLogin — учётные данные для браузерного POST-логина в ejudge master.
type BrowserLogin struct {
	BaseURL   string
	Login     string
	Password  string
	ContestID int
}

type CredentialsRepository interface {
	Get(ctx context.Context, login string) (Credentials, bool, error)
	Upsert(ctx context.Context, cred Credentials) error
}

type Gateway struct {
	teachers    teacherRepository
	credentials CredentialsRepository

	baseURL string
	timeout time.Duration

	mu sync.Mutex
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

	g.mu.Lock()
	defer g.mu.Unlock()

	cred, found, err := g.credentials.Get(ctx, login)
	if err != nil {
		return nil, err
	}
	if !found || cred.APIKey == "" {
		if err := g.ensureAPIKey(ctx, login); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrAPIKeyProvision, err)
		}
	}
	return g.clientWithAPIKey(ctx, login)
}

// EnsureAPIKey создаёт ejudge API key, если его ещё нет.
func (g *Gateway) EnsureAPIKey(ctx context.Context) error {
	if g.baseURL == "" {
		return fmt.Errorf("ejudge gateway: base URL not configured")
	}

	login, ok := auth.LoginFrom(ctx)
	if !ok || login == "" {
		return ErrNoLogin
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	cred, found, err := g.credentials.Get(ctx, login)
	if err != nil {
		return err
	}
	if !found || cred.APIKey == "" {
		return g.ensureAPIKey(ctx, login)
	}
	return nil
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

func (g *Gateway) BrowserLogin(ctx context.Context, contestID int) (BrowserLogin, bool, error) {
	if g.baseURL == "" || contestID <= 0 {
		return BrowserLogin{}, false, nil
	}
	login, hasLogin := auth.LoginFrom(ctx)
	if !hasLogin || login == "" {
		return BrowserLogin{}, false, nil
	}

	rec, ok, err := g.teachers.Get(ctx, login)
	if err != nil {
		return BrowserLogin{}, false, err
	}
	if !ok {
		return BrowserLogin{}, false, nil
	}

	return BrowserLogin{
		BaseURL:   g.baseURL,
		Login:     login,
		Password:  rec.Password,
		ContestID: contestID,
	}, true, nil
}

func (g *Gateway) ListContests(ctx context.Context) ([]servecontrol.Brief, error) {
	if g.baseURL == "" {
		return nil, fmt.Errorf("ejudge gateway: base URL not configured")
	}
	login, ok := auth.LoginFrom(ctx)
	if !ok || login == "" {
		return nil, ErrNoLogin
	}
	rec, found, err := g.teachers.Get(ctx, login)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrTeacherNotFound
	}

	sess, err := ejauth.ServeControlSessionLogin(ctx, g.baseURL, login, rec.Password, g.timeout)
	if err != nil {
		return nil, err
	}
	return servecontrol.ListContests(ctx, sess, g.timeout)
}

func (g *Gateway) ensureAPIKey(ctx context.Context, login string) error {
	rec, ok, err := g.teachers.Get(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTeacherNotFound
	}

	apiKey, err := ejauth.IssueAPIKey(ctx, g.baseURL, login, rec.Password)
	if err != nil {
		return err
	}

	return g.credentials.Upsert(ctx, Credentials{
		Login:  login,
		APIKey: apiKey,
	})
}
