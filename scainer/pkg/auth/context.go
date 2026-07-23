package auth

import (
	"context"
)

type ctxKey int

const loginKey ctxKey = iota

func WithLogin(ctx context.Context, login string) context.Context {
	return context.WithValue(ctx, loginKey, login)
}

func LoginFrom(ctx context.Context) (string, bool) {
	login, ok := ctx.Value(loginKey).(string)
	return login, ok && login != ""
}
