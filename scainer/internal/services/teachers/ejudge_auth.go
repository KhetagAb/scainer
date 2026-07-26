package teachers

import (
	"context"
	"time"

	ejauth "scainer/pkg/ejudge/auth"
)

type EjudgeAuth struct {
	BaseURL string
	Timeout time.Duration
}

func (a EjudgeAuth) Validate(ctx context.Context, login, password string) error {
	_, err := ejauth.MasterSessionLogin(ctx, a.BaseURL, login, password)
	return err
}
