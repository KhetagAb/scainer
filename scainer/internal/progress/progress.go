package progress

import "context"

type Event struct {
	Phase string // "importing" | "analyzing"
	Done  int
	Total int
	Label string
}

type Reporter func(Event)

type key struct{}

func With(ctx context.Context, r Reporter) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, key{}, r)
}

func Report(ctx context.Context, e Event) {
	if r, ok := ctx.Value(key{}).(Reporter); ok && r != nil {
		r(e)
	}
}
