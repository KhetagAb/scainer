package analyze

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
)

type InvalidationPolicy[U domain.Unit] interface {
	ScopeKey(U) contests.ScopeKey
	ShouldRun(ctx context.Context, u U, prev contests.DetectorProgress) bool
	AfterRun(u U, next *contests.DetectorProgress)
}

type EveryRun[U domain.Unit] struct {
	Key func(U) contests.ScopeKey
}

func (EveryRun[U]) ShouldRun(context.Context, U, contests.DetectorProgress) bool {
	return true
}

func (EveryRun[U]) AfterRun(U, *contests.DetectorProgress) {}

func (p EveryRun[U]) ScopeKey(u U) contests.ScopeKey {
	return p.Key(u)
}

type manualKey struct{}

func WithManual(ctx context.Context) context.Context {
	return context.WithValue(ctx, manualKey{}, true)
}

func ManualFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(manualKey{}).(bool)
	return v
}
